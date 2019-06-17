package iofog

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

// LocalContainer struct to encapsulate utilities around docker
type LocalContainer struct {
	client *client.Client
}

type port struct {
	Host      string
	Container *LocalContainerPort
}

type LocalControllerConfig struct {
	Host           string
	ContainerNames map[string]string
	ControllerPort port
	ConnectorPort  port
}

type LocalContainerPort struct {
	Protocol string
	Port     string
}

type LocalUserConfig struct {
	config.IofogUser
}

type LocalAgentConfig struct {
	Host          string
	AgentPort     port
	ContainerName string
	Name          string
}

// NewAgentConfig generates a static agent config
func NewLocalAgentConfig(name string) *LocalAgentConfig {
	return &LocalAgentConfig{
		Host:          "0.0.0.0",
		AgentPort:     port{Host: "54321", Container: &LocalContainerPort{Protocol: "tcp", Port: "54321"}},
		ContainerName: fmt.Sprintf("iofog-agent-%s", name),
		Name:          name,
	}
}

// NewLocalControllerConfig generats a static controller config
func NewLocalControllerConfig(name string) *LocalControllerConfig {
	nameMap := make(map[string]string)
	nameMap["connector"] = "iofog-connector-" + name
	nameMap["controller"] = "iofog-controller-" + name
	return &LocalControllerConfig{
		Host:           "0.0.0.0",
		ContainerNames: nameMap,
		ControllerPort: port{Host: "51121", Container: &LocalContainerPort{Port: "51121", Protocol: "tcp"}},
		ConnectorPort:  port{Host: "53321", Container: &LocalContainerPort{Port: "8080", Protocol: "tcp"}},
	}
}

// GetLocalUserConfig return the user config
func GetLocalUserConfig(namespace string, controllerName string) *LocalUserConfig {
	ctrl, err := config.GetController(namespace, controllerName)
	if err == nil {
		// Use existing user
		return &LocalUserConfig{ctrl.IofogUser}
	} else {
		// Generate new user
		return &LocalUserConfig{
			config.NewUser(),
		}
	}
}

// NewLocalContainerClient returns a LocalContainer struct
func NewLocalContainerClient() (*LocalContainer, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	return &LocalContainer{
		client: cli,
	}, nil
}

func (lc *LocalContainer) getContainerByName(name string) (types.Container, error) {
	ctx := context.Background()
	containers, err := lc.client.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return types.Container{}, err
	}

	for _, container := range containers {
		for _, containerName := range container.Names {
			if containerName == "/"+name { // Docker prefixes names with /
				return container, nil
			}
		}
	}
	return types.Container{}, util.NewInputError(fmt.Sprintf("Could not find container %s", name))
}

// CleanContainer stops and remove a container based on a container name
func (lc *LocalContainer) CleanContainer(name string) error {
	ctx := context.Background()

	container, err := lc.getContainerByName(name)
	if err != nil {
		return err
	}
	if err := lc.client.ContainerStop(ctx, container.ID, nil); err != nil {
		return err
	}
	return lc.client.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{Force: true})
}

// DeployContainer deploys a container based on an image and a port mappin
func (lc *LocalContainer) DeployContainer(image, name string, ports map[string]*LocalContainerPort) (string, error) {
	ctx := context.Background()

	portSet := nat.PortSet{}
	portMap := nat.PortMap{}

	for hostPort, containerPort := range ports {
		natPort, err := nat.NewPort(containerPort.Protocol, containerPort.Port)
		if err != nil {
			return "", err
		}
		portSet[natPort] = struct{}{}
		portMap[natPort] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: hostPort,
			},
		}
	}

	containerConfig := &dockerContainer.Config{
		Image:        image,
		ExposedPorts: portSet,
	}
	hostConfig := &dockerContainer.HostConfig{
		PortBindings: portMap,
	}

	out, err := lc.client.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return "", err
	}
	io.Copy(os.Stdout, out)

	container, err := lc.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, name)
	if err != nil {
		return "", err
	}

	return container.ID, lc.client.ContainerStart(ctx, container.ID, types.ContainerStartOptions{})
}

func (lc *LocalContainer) ExecuteCmd(name string, cmd []string) (err error) {
	ctx := context.Background()

	container, err := lc.getContainerByName(name)
	if err != nil {
		return err
	}

	execConfig := types.ExecConfig{AttachStdout: true, AttachStderr: true,
		Cmd: cmd}

	execID, err := lc.client.ContainerExecCreate(ctx, container.ID, execConfig)
	if err != nil {
		return err
	}

	res, err := lc.client.ContainerExecAttach(ctx, execID.ID, execConfig)
	if err != nil {
		return err
	}
	defer res.Close()

	return lc.client.ContainerExecStart(ctx, execID.ID, types.ExecStartCheck{})
}
