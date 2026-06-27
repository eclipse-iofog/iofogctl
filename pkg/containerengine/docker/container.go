package docker

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	mobyclient "github.com/moby/moby/client"
)

const (
	RestartPolicyAlways = "always"
	NetworkModeHost     = "host"
)

// ContainerSummary is a minimal container listing entry.
type ContainerSummary struct {
	ID     string
	Names  []string
	Image  string
	Status string
	State  string
}

// DeployOptions configures container create/start.
type DeployOptions struct {
	Name          string
	Image         string
	Env           []string
	Binds         []string
	Privileged    bool
	NetworkMode   string
	PortBindings  map[string]PortBinding // hostPort -> container port/proto
	RestartPolicy string
	StopTimeout   time.Duration
}

// PortBinding maps a container port to a host port.
type PortBinding struct {
	HostIP        string
	HostPort      string
	ContainerPort string
	Protocol      string
}

// ExecResult holds output from a container exec.
type ExecResult struct {
	StdOut   string
	StdErr   string
	ExitCode int
}

// ListContainers returns containers; when all is false, only running containers are returned.
func (c *Client) ListContainers(all bool) ([]ContainerSummary, error) {
	result, err := c.cli.ContainerList(c.context(), mobyclient.ContainerListOptions{All: all})
	if err != nil {
		return nil, err
	}
	out := make([]ContainerSummary, 0, len(result.Items))
	for _, item := range result.Items {
		out = append(out, ContainerSummary{
			ID:     item.ID,
			Names:  item.Names,
			Image:  item.Image,
			Status: item.Status,
			State:  string(item.State),
		})
	}
	return out, nil
}

// GetContainerByName finds a container by exact name (with or without leading slash).
func (c *Client) GetContainerByName(name string) (ContainerSummary, error) {
	containers, err := c.ListContainers(true)
	if err != nil {
		return ContainerSummary{}, err
	}
	target := normalizeContainerName(name)
	for _, cont := range containers {
		for _, containerName := range cont.Names {
			if normalizeContainerName(containerName) == target {
				return cont, nil
			}
		}
	}
	return ContainerSummary{}, util.NewInputError(fmt.Sprintf("Could not find container %s", name))
}

func normalizeContainerName(name string) string {
	return strings.TrimPrefix(strings.TrimSpace(name), "/")
}

// RemoveContainerByName stops and force-removes a container if it exists.
func (c *Client) RemoveContainerByName(name string) error {
	cont, err := c.GetContainerByName(name)
	if err != nil {
		var inputErr *util.InputError
		if errors.As(err, &inputErr) {
			return nil
		}
		return err
	}
	return c.RemoveContainerByID(cont.ID)
}

// RemoveContainerByID stops and force-removes a container by ID.
func (c *Client) RemoveContainerByID(id string) error {
	ctx := c.context()
	timeout := 60
	_, _ = c.cli.ContainerStop(ctx, id, mobyclient.ContainerStopOptions{Timeout: &timeout})
	_, err := c.cli.ContainerRemove(ctx, id, mobyclient.ContainerRemoveOptions{Force: true})
	return err
}

// DeployContainer pulls (when needed), recreates, and starts a container.
func (c *Client) DeployContainer(opts DeployOptions, pullOpts PullOptions) (string, error) {
	if err := c.RemoveContainerByName(opts.Name); err != nil {
		return "", err
	}
	if err := c.PullImage(opts.Image, pullOpts); err != nil {
		return "", err
	}

	portSet := network.PortSet{}
	portMap := network.PortMap{}
	for hostPort, binding := range opts.PortBindings {
		proto := network.TCP
		if strings.EqualFold(binding.Protocol, "udp") {
			proto = network.UDP
		}
		containerPort := binding.ContainerPort
		if containerPort == "" {
			containerPort = hostPort
		}
		portNum, err := strconv.ParseUint(containerPort, 10, 16)
		if err != nil {
			return "", fmt.Errorf("invalid container port %q: %w", containerPort, err)
		}
		natPort, ok := network.PortFrom(uint16(portNum), proto)
		if !ok {
			return "", fmt.Errorf("invalid container port %q", containerPort)
		}
		portSet[natPort] = struct{}{}
		hostBindingPort := binding.HostPort
		if hostBindingPort == "" {
			hostBindingPort = hostPort
		}
		portMap[natPort] = []network.PortBinding{{HostPort: hostBindingPort}}
	}

	networkMode := container.NetworkMode(opts.NetworkMode)
	restartName := container.RestartPolicyAlways
	if opts.RestartPolicy != "" {
		restartName = container.RestartPolicyMode(opts.RestartPolicy)
	}
	restartPolicy := container.RestartPolicy{Name: restartName}

	hostConfig := &container.HostConfig{
		Binds:         opts.Binds,
		Privileged:    opts.Privileged,
		NetworkMode:   networkMode,
		PortBindings:  portMap,
		RestartPolicy: restartPolicy,
	}

	containerConfig := &container.Config{
		Image:        opts.Image,
		Env:          opts.Env,
		ExposedPorts: portSet,
	}
	if stopTimeout := int(opts.StopTimeout.Seconds()); stopTimeout > 0 {
		containerConfig.StopTimeout = &stopTimeout
	}

	createResp, err := c.cli.ContainerCreate(c.context(), mobyclient.ContainerCreateOptions{
		Name:       opts.Name,
		Config:     containerConfig,
		HostConfig: hostConfig,
	})
	if err != nil {
		return "", util.NewError(fmt.Sprintf("Failed to create container: %v", err))
	}

	_, err = c.cli.ContainerStart(c.context(), createResp.ID, mobyclient.ContainerStartOptions{})
	if err != nil {
		return "", util.NewError(fmt.Sprintf("Failed to start container: %v", err))
	}
	return createResp.ID, nil
}

// GetContainerIP returns the primary IPv4 address for a container.
func (c *Client) GetContainerIP(name string) (string, error) {
	cont, err := c.GetContainerByName(name)
	if err != nil {
		return "", err
	}
	inspect, err := c.cli.ContainerInspect(c.context(), cont.ID, mobyclient.ContainerInspectOptions{})
	if err != nil {
		return "", err
	}
	cfg := inspect.Container
	if cfg.HostConfig != nil && cfg.HostConfig.NetworkMode == NetworkModeHost {
		return "127.0.0.1", nil
	}
	if cfg.NetworkSettings == nil {
		return "", util.NewNotFoundError(fmt.Sprintf("Container %s: no network settings", name))
	}
	mode := "bridge"
	if cfg.HostConfig != nil && cfg.HostConfig.NetworkMode != "" {
		mode = string(cfg.HostConfig.NetworkMode)
	}
	if net, ok := cfg.NetworkSettings.Networks[mode]; ok && net.IPAddress.IsValid() {
		return net.IPAddress.String(), nil
	}
	for _, candidate := range cfg.NetworkSettings.Networks {
		if candidate != nil && candidate.IPAddress.IsValid() {
			return candidate.IPAddress.String(), nil
		}
	}
	return "", util.NewNotFoundError(fmt.Sprintf("Container %s: could not find network setting for network %s", name, mode))
}

// GetLogsByName returns stdout and stderr log streams for a container.
func (c *Client) GetLogsByName(name string) (stdout, stderr string, err error) {
	cont, err := c.GetContainerByName(name)
	if err != nil {
		return "", "", err
	}
	logs, err := c.cli.ContainerLogs(c.context(), cont.ID, mobyclient.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		return "", "", err
	}
	defer logs.Close()

	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)
	if _, err = stdcopy.StdCopy(stdoutBuf, stderrBuf, logs); err != nil {
		return "", "", err
	}
	return stdoutBuf.String(), stderrBuf.String(), nil
}

// ExecuteCmd runs a command inside a container and returns its output.
func (c *Client) ExecuteCmd(name string, cmd []string) (ExecResult, error) {
	var result ExecResult
	cont, err := c.GetContainerByName(name)
	if err != nil {
		return result, err
	}

	ctx := c.context()
	execResp, err := c.cli.ExecCreate(ctx, cont.ID, mobyclient.ExecCreateOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return result, err
	}

	attachResp, err := c.cli.ExecAttach(ctx, execResp.ID, mobyclient.ExecAttachOptions{})
	if err != nil {
		return result, err
	}
	defer attachResp.Close()

	var outBuf, errBuf bytes.Buffer
	if _, err = stdcopy.StdCopy(&outBuf, &errBuf, attachResp.Reader); err != nil && !errors.Is(err, io.EOF) {
		return result, err
	}

	inspect, err := c.cli.ExecInspect(ctx, execResp.ID, mobyclient.ExecInspectOptions{})
	if err != nil {
		return result, err
	}

	result.ExitCode = inspect.ExitCode
	result.StdOut = outBuf.String()
	result.StdErr = errBuf.String()
	return result, nil
}
