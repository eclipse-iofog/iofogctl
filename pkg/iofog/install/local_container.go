/*
 *  *******************************************************************************
 *  * Copyright (c) 2019 Edgeworx, Inc.
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package install

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

// LocalContainer struct to encapsulate utilities around docker
type LocalContainer struct {
	client *client.Client
}

// ExecResult contains the output of a command ran into docker exec
type ExecResult struct {
	StdOut   string
	StdErr   string
	ExitCode int
}

type port struct {
	Host      string
	Container *LocalContainerPort
}

type Credentials struct {
	User     string
	Password string
}

type LocalContainerConfig struct {
	Host          string
	Ports         []port
	ContainerName string
	Image         string
	DefaultImage  string
	Privileged    bool
	Binds         []string
	NetworkMode   string
	Credentials   Credentials
}

type LocalControllerConfig struct {
	ContainerMap map[string]*LocalContainerConfig
}

type LocalContainerPort struct {
	Protocol string
	Port     string
}

type LocalAgentConfig struct {
	LocalContainerConfig
	Name string
}

func GetLocalContainerName(t string) string {
	names := map[string]string{
		"controller": sanitizeContainerName("iofog-controller"),
		"connector":  sanitizeContainerName("iofog-connector"),
		"agent":      sanitizeContainerName("iofog-agent"),
	}
	name, ok := names[t]
	if ok == false {
		return ""
	}
	return name
}

func sanitizeContainerName(name string) string {
	r := regexp.MustCompile("[^a-zA-Z0-9_.-]")
	return r.ReplaceAllString(name, "-")
}

// NewAgentConfig generates a static agent config
func NewLocalAgentConfig(name string, image string, ctrlConfig *LocalContainerConfig, credentials Credentials) *LocalAgentConfig {
	if image == "" {
		image = "docker.io/iofog/agent:" + util.GetAgentTag()
	}

	var bindings []string

	if runtime.GOOS == "windows" {
		bindings = append(bindings, "//var/run/docker.sock:/var/run/docker.sock:rw")
	} else {
		bindings = append(bindings, "/var/run/docker.sock:/var/run/docker.sock:rw")
	}

	return &LocalAgentConfig{
		LocalContainerConfig: LocalContainerConfig{
			Host: "0.0.0.0",
			Ports: []port{
				{Host: "54321", Container: &LocalContainerPort{Protocol: "tcp", Port: "54321"}},
				{Host: "8081", Container: &LocalContainerPort{Protocol: "tcp", Port: "22"}},
			},
			ContainerName: GetLocalContainerName("agent"),
			Image:         image,
			Privileged:    true,
			Binds:         []string{"/var/run/docker.sock:/var/run/docker.sock:rw"},
			NetworkMode:   "bridge",
			Credentials:   credentials,
		},
		Name: name,
	}
}

// NewLocalConnectorConfig generates a static connector config
func NewLocalConnectorConfig(image string, credentials Credentials) *LocalContainerConfig {
	if image == "" {
		image = "docker.io/iofog/connector:" + util.GetConnectorTag()
	}

	return &LocalContainerConfig{
		Host:          "0.0.0.0",
		Ports:         []port{{Host: iofog.ConnectorPortString, Container: &LocalContainerPort{Port: iofog.ConnectorPortString, Protocol: "tcp"}}},
		ContainerName: GetLocalContainerName("connector"),
		Image:         image,
		Privileged:    false,
		Binds:         []string{},
		NetworkMode:   "bridge",
		Credentials:   credentials,
	}

}

// NewLocalControllerConfig generats a static controller config
func NewLocalControllerConfig(image string, credentials Credentials) *LocalContainerConfig {
	if image == "" {
		image = "docker.io/iofog/controller:" + util.GetControllerTag()
	}
	return &LocalContainerConfig{
		Host: "0.0.0.0",
		Ports: []port{
			{Host: iofog.ControllerPortString, Container: &LocalContainerPort{Port: iofog.ControllerPortString, Protocol: "tcp"}},
			{Host: iofog.ControllerHostECNViewerPortString, Container: &LocalContainerPort{Port: iofog.DefaultHTTPPortString, Protocol: "tcp"}},
		},
		ContainerName: GetLocalContainerName("controller"),
		Image:         image,
		Privileged:    false,
		Binds:         []string{},
		NetworkMode:   "bridge",
		Credentials:   credentials,
	}
}

// NewLocalContainerClient returns a LocalContainer struct
func NewLocalContainerClient() (*LocalContainer, error) {
	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	if err = client.FromEnv(cli); err != nil {
		return nil, err
	}
	return &LocalContainer{
		client: cli,
	}, nil
}

// GetLogsByName returns the logs of the container specified by name
func (lc *LocalContainer) GetLogsByName(name string) (stdout, stderr string, err error) {
	ctx := context.Background()
	r, err := lc.client.ContainerLogs(ctx, name, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return
	}
	defer r.Close()

	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	_, err = stdcopy.StdCopy(stdoutBuf, stderrBuf, r)
	if err != nil {
		return
	}

	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()

	return
}

func (lc *LocalContainer) GetContainerByName(name string) (types.Container, error) {
	ctx := context.Background()
	// List containers
	containers, err := lc.client.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return types.Container{}, err
	}

	// Find by name
	for _, container := range containers {
		for _, containerName := range container.Names {
			if containerName == "/"+name { // Docker prefixes names with /
				return container, nil
			}
		}
	}
	return types.Container{}, util.NewInputError(fmt.Sprintf("Could not find container %s", name))
}

func (lc *LocalContainer) ListContainers() ([]types.Container, error) {
	ctx := context.Background()
	return lc.client.ContainerList(ctx, types.ContainerListOptions{})
}

// CleanContainer stops and remove a container based on a container name
func (lc *LocalContainer) CleanContainer(name string) error {
	ctx := context.Background()

	container, err := lc.GetContainerByName(name)
	if err != nil {
		return err
	}
	// Stop container if running (ignore error if there is no running container)
	lc.client.ContainerStop(ctx, container.ID, nil)

	// Force remove container
	return lc.client.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{Force: true})
}

func (lc *LocalContainer) CleanContainerByID(id string) error {
	ctx := context.Background()

	// Stop container if running (ignore error if there is no running container)
	lc.client.ContainerStop(ctx, id, nil)

	// Force remove container
	return lc.client.ContainerRemove(ctx, id, types.ContainerRemoveOptions{Force: true})
}

func (lc *LocalContainer) getPullOptions(config *LocalContainerConfig) (ret types.ImagePullOptions) {
	dockerUser := config.Credentials.User
	dockerPwd := config.Credentials.Password

	if dockerUser != "" {
		authConfig := types.AuthConfig{
			Username: dockerUser,
			Password: dockerPwd,
		}
		encodedJSON, err := json.Marshal(authConfig)
		if err != nil {
			panic(err)
		}
		authStr := base64.URLEncoding.EncodeToString(encodedJSON)
		ret.RegistryAuth = authStr
	}
	return
}

func getImageTag(image string) string {
	if strings.HasPrefix(image, "docker.io/") {
		return image[len("docker.io/"):]
	}
	return image
}

func (lc *LocalContainer) waitForImage(image string, counter int8) error {
	if counter >= 18 { // 180 seconds
		return util.NewInternalError("Could not find newly pulled image: " + image)
	}
	ctx := context.Background()
	imgs, listErr := lc.client.ImageList(ctx, types.ImageListOptions{All: true})
	if listErr != nil {
		fmt.Printf("Could not list local images: %v\n", listErr)
		return listErr
	}
	for idx := range imgs {
		for _, tag := range imgs[idx].RepoTags {
			if tag == image {
				return nil
			}
		}
	}
	time.Sleep(10 * time.Second)
	return lc.waitForImage(image, counter+1)
}

// DeployContainer deploys a container based on an image and a port mappin
func (lc *LocalContainer) DeployContainer(containerConfig *LocalContainerConfig) (string, error) {
	ctx := context.Background()

	portSet := nat.PortSet{}
	portMap := nat.PortMap{}

	// Create port mappings
	for _, port := range containerConfig.Ports {
		natPort, err := nat.NewPort(port.Container.Protocol, port.Container.Port)
		if err != nil {
			return "", err
		}
		portSet[natPort] = struct{}{}
		portMap[natPort] = []nat.PortBinding{
			{
				HostIP:   containerConfig.Host,
				HostPort: port.Host,
			},
		}
	}

	dockerContainerConfig := &dockerContainer.Config{
		Image:        containerConfig.Image,
		ExposedPorts: portSet,
	}
	hostConfig := &dockerContainer.HostConfig{
		PortBindings: portMap,
		Privileged:   containerConfig.Privileged,
		Binds:        containerConfig.Binds,
		NetworkMode:  dockerContainer.NetworkMode(containerConfig.NetworkMode),
	}

	// Pull image
	reader, err := lc.client.ImagePull(ctx, containerConfig.Image, lc.getPullOptions(containerConfig))
	imageTag := getImageTag(containerConfig.Image)
	if err != nil {
		fmt.Printf("Could not pull image: %v, listing local images...\n", err)
		imgs, listErr := lc.client.ImageList(ctx, types.ImageListOptions{All: true})
		if listErr != nil {
			fmt.Printf("Could not list local images: %v\n", listErr)
			return "", err
		}
		found := false
		for idx := range imgs {
			for _, tag := range imgs[idx].RepoTags {
				if tag == imageTag {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			fmt.Printf("Could not pull image: %v\n Could not find image [%v] locally, please run docker pull [%v]\n", err, containerConfig.Image, containerConfig.Image)
			return "", err
		}
	} else {
		defer reader.Close()
		_, err := ioutil.ReadAll(reader)
		if err != nil {
			return "", err
		}
		// Wait for image to be discoverable by docker daemon
		err = lc.waitForImage(imageTag, 0)
		if err != nil {
			return "", err
		}
	}

	// Create network if it does not exists
	// networkName := "local-iofog-network"
	// networks, err := lc.client.NetworkList(ctx, types.NetworkListOptions{})
	// networkID := ""
	// for i := range networks {
	// 	if networks[i].Name == networkName {
	// 		networkID = networks[i].ID
	// 		break
	// 	}
	// }

	// if networkID == "" {
	// 	networkResponse, err := lc.client.NetworkCreate(ctx, networkName, types.NetworkCreate{
	// 		Driver:         "bridge",
	// 		CheckDuplicate: true,
	// 	})
	// 	if err != nil {
	// 		fmt.Printf("Failed to create network: %v\n", err)
	// 		return "", err
	// 	}
	// 	networkID = networkResponse.ID
	// }

	container, err := lc.client.ContainerCreate(ctx, dockerContainerConfig, hostConfig, nil, containerConfig.ContainerName)
	if err != nil {
		fmt.Printf("Failed to create container: %v\n", err)
		return "", err
	}

	// Connect to network
	// err = lc.client.NetworkConnect(ctx, networkID, container.ID, nil)
	// if err != nil {
	// 	fmt.Printf("Failed to connect container to network: %v\n", err)
	// 	return "", err
	// }

	// Start container
	err = lc.client.ContainerStart(ctx, container.ID, types.ContainerStartOptions{})
	if err != nil {
		fmt.Printf("Failed to start container: %v\n", err)
		return "", err
	}

	return container.ID, err
}

// Returns endpoint to reach controller container from within another container
func (lc *LocalContainer) GetLocalControllerEndpoint() (controllerEndpoint string, err error) {
	host, err := lc.GetContainerIP(GetLocalContainerName("controller"))
	if err != nil {
		return controllerEndpoint, err
	}
	controllerEndpoint = fmt.Sprintf("%s:%s", host, iofog.ControllerPortString)
	return
}

func (lc *LocalContainer) GetContainerIP(name string) (IP string, err error) {
	container, err := lc.GetContainerByName(name)
	if err != nil {
		return
	}

	network, found := container.NetworkSettings.Networks[container.HostConfig.NetworkMode]
	if found == false {
		return "", util.NewNotFoundError(fmt.Sprintf("Container %s : Could not find network setting for network %s", name, container.HostConfig.NetworkMode))
	}

	return network.IPAddress, nil
}

func (lc *LocalContainer) WaitForCommand(containerName string, condition *regexp.Regexp, command ...string) error {
	for iteration := 0; iteration < 30; iteration++ {
		output, err := lc.ExecuteCmd(containerName, command)
		if err != nil {
			Verbose(fmt.Sprintf("Container command %v failed with error %v\n", command, err.Error()))
		}
		if condition.MatchString(output.StdOut) {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return util.NewInternalError("Timed out waiting for container")
}

func (lc *LocalContainer) ExecuteCmd(name string, cmd []string) (execResult ExecResult, err error) {
	ctx := context.Background()

	container, err := lc.GetContainerByName(name)
	if err != nil {
		return
	}

	// Create command to execute inside container
	execConfig := types.ExecConfig{AttachStdout: true, AttachStderr: true,
		Cmd: cmd}
	execStartCheck := types.ExecStartCheck{}

	execID, err := lc.client.ContainerExecCreate(ctx, container.ID, execConfig)
	if err != nil {
		return
	}

	// Attach command to container
	res, err := lc.client.ContainerExecAttach(ctx, execID.ID, execStartCheck)
	if err != nil {
		return
	}
	defer res.Close()

	// read the output
	var outBuf, errBuf bytes.Buffer
	outputDone := make(chan error)

	go func() {
		// StdCopy demultiplexes the stream into two buffers
		_, err = stdcopy.StdCopy(&outBuf, &errBuf, res.Reader)
		outputDone <- err
	}()

	select {
	case err := <-outputDone:
		if err != nil {
			return execResult, err
		}
		break

	case <-ctx.Done():
		return execResult, ctx.Err()
	}

	stdout, err := ioutil.ReadAll(&outBuf)
	if err != nil {
		return execResult, err
	}
	stderr, err := ioutil.ReadAll(&errBuf)
	if err != nil {
		return execResult, err
	}

	inspect, err := lc.client.ContainerExecInspect(ctx, execID.ID)
	if err != nil {
		return execResult, err
	}

	execResult.ExitCode = inspect.ExitCode
	execResult.StdOut = string(stdout)
	execResult.StdErr = string(stderr)

	// Run command
	if err = lc.client.ContainerExecStart(ctx, execID.ID, execStartCheck); err != nil {
		return
	}
	return
}
