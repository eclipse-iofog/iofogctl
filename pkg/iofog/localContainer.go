package iofog

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"time"

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
	DefaultImages  map[string]string
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
	DefaultImage  string
}

// NewAgentConfig generates a static agent config
func NewLocalAgentConfig(name string) *LocalAgentConfig {
	return &LocalAgentConfig{
		Host:          "0.0.0.0",
		AgentPort:     port{Host: "54321", Container: &LocalContainerPort{Protocol: "tcp", Port: "54321"}},
		ContainerName: fmt.Sprintf("iofog-agent-%s", name),
		Name:          name,
		DefaultImage:  "docker.io/iofog/agent",
	}
}

// NewLocalControllerConfig generats a static controller config
func NewLocalControllerConfig(name string) *LocalControllerConfig {
	nameMap := make(map[string]string)
	nameMap["connector"] = "iofog-connector-" + name
	nameMap["controller"] = "iofog-controller-" + name

	imageMap := make(map[string]string)
	imageMap["connector"] = "docker.io/iofog/connector"
	imageMap["controller"] = "docker.io/iofog/controller"
	return &LocalControllerConfig{
		Host:           "0.0.0.0",
		ContainerNames: nameMap,
		ControllerPort: port{Host: "51121", Container: &LocalContainerPort{Port: "51121", Protocol: "tcp"}},
		ConnectorPort:  port{Host: "8080", Container: &LocalContainerPort{Port: "8080", Protocol: "tcp"}},
		DefaultImages:  imageMap,
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
			config.NewRandomUser(),
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

// CleanContainer stops and remove a container based on a container name
func (lc *LocalContainer) CleanContainer(name string) error {
	ctx := context.Background()

	container, err := lc.getContainerByName(name)
	if err != nil {
		return err
	}
	// Stop container
	if err := lc.client.ContainerStop(ctx, container.ID, nil); err != nil {
		return err
	}
	// Force remove container
	return lc.client.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{Force: true})
}

func (lc *LocalContainer) getPullOptions(image string) (ret types.ImagePullOptions) {
	dockerUser := ""
	dockerPwd := ""
	// TODO: AlexD - Find a more elegant way to specify docker auth. (if needed)
	gcrRegex := regexp.MustCompile("^((us|eu|asia)\\.){0,1}gcr\\.io\\/")
	if gcrRegex.MatchString(image) {
		dockerUser = "_json_key"
		out, err := exec.Command("cat", "./edgeworx-iofog-95aff71cbc7a.json").Output()
		if err != nil {
			fmt.Printf("Failed to get gcloud auth token: %v\n", err)
			return
		}
		dockerPwd = string(out)
	}

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

// DeployContainer deploys a container based on an image and a port mappin
func (lc *LocalContainer) DeployContainer(image, name string, ports map[string]*LocalContainerPort) (string, error) {
	ctx := context.Background()

	portSet := nat.PortSet{}
	portMap := nat.PortMap{}

	// Create port mappings
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

	// Pull image
	_, err := lc.client.ImagePull(ctx, image, lc.getPullOptions(image))
	if err != nil {
		return "", err
	}

	// Create container
	container, err := lc.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, name)
	if err != nil {
		return "", err
	}

	// Start container
	return container.ID, lc.client.ContainerStart(ctx, container.ID, types.ContainerStartOptions{})
}

func (lc *LocalContainer) WaitForCommand(condition *regexp.Regexp, command string, args ...string) error {
	for iteration := 0; iteration < 30; iteration++ {
		output, _ := exec.Command(command, args...).Output()
		if condition.MatchString(string(output)) {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return util.NewInternalError("Timed out waiting for container")
}

func (lc *LocalContainer) ExecuteCmd(name string, cmd []string) (err error) {
	ctx := context.Background()

	container, err := lc.getContainerByName(name)
	if err != nil {
		return err
	}

	// Create command to execute inside container
	execConfig := types.ExecConfig{AttachStdout: true, AttachStderr: true,
		Cmd: cmd}

	execID, err := lc.client.ContainerExecCreate(ctx, container.ID, execConfig)
	if err != nil {
		return err
	}

	// Attach command to container
	res, err := lc.client.ContainerExecAttach(ctx, execID.ID, execConfig)
	if err != nil {
		return err
	}
	defer res.Close()

	// Run command
	return lc.client.ContainerExecStart(ctx, execID.ID, types.ExecStartCheck{})
}
