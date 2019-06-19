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

type LocalContainerConfig struct {
	Host          string
	Ports         []port
	ContainerName string
	Image         string
	DefaultImage  string
	Privileged    bool
	Binds         []string
	NetworkMode   string
	Links         []string
}

type LocalControllerConfig struct {
	Name         string
	ContainerMap map[string]*LocalContainerConfig
}

type LocalContainerPort struct {
	Protocol string
	Port     string
}

type LocalUserConfig struct {
	config.IofogUser
}

type LocalAgentConfig struct {
	LocalContainerConfig
	Name string
}

func sanitizeContainerName(name string) string {
	r := regexp.MustCompile("[^a-zA-Z0-9_.-]")
	return r.ReplaceAllString(name, "-")
}

// NewAgentConfig generates a static agent config
func NewLocalAgentConfig(name string, image string, ctrlConfig *LocalContainerConfig) *LocalAgentConfig {
	if image == "" {
		image = "docker.io/iofog/agent"
	}
	return &LocalAgentConfig{
		LocalContainerConfig: LocalContainerConfig{
			Host: "0.0.0.0",
			Ports: []port{
				port{Host: "54321", Container: &LocalContainerPort{Protocol: "tcp", Port: "54321"}},
				port{Host: "8081", Container: &LocalContainerPort{Protocol: "tcp", Port: "22"}},
			},
			ContainerName: sanitizeContainerName(fmt.Sprintf("iofog-agent-%s", name)),
			Image:         image,
			Privileged:    true,
			Binds:         []string{"/var/run/docker.sock:/var/run/docker.sock:rw"},
			NetworkMode:   "bridge",
			Links:         []string{fmt.Sprintf("%s:%s", ctrlConfig.ContainerName, ctrlConfig.ContainerName)},
		},
		Name: name,
	}
}

// NewLocalControllerConfig generats a static controller config
func NewLocalControllerConfig(name string, images map[string]string) *LocalControllerConfig {
	controllerImg, exists := images["controller"]
	if !exists {
		controllerImg = "docker.io/iofog/controller"
	}
	containerMap := make(map[string]*LocalContainerConfig)
	containerMap["controller"] = &LocalContainerConfig{
		Host:          "0.0.0.0",
		Ports:         []port{port{Host: "51121", Container: &LocalContainerPort{Port: "51121", Protocol: "tcp"}}},
		ContainerName: sanitizeContainerName("iofog-controller-" + name),
		Image:         controllerImg,
		Privileged:    false,
		Binds:         []string{},
		NetworkMode:   "bridge",
	}

	connectorImg, exists := images["connector"]
	if !exists {
		connectorImg = "docker.io/iofog/connector"
	}

	containerMap["connector"] = &LocalContainerConfig{
		Host:          "0.0.0.0",
		Ports:         []port{port{Host: "8080", Container: &LocalContainerPort{Port: "8080", Protocol: "tcp"}}},
		ContainerName: sanitizeContainerName("iofog-connector-" + name),
		Image:         connectorImg,
		Privileged:    false,
		Binds:         []string{},
		NetworkMode:   "bridge",
	}

	return &LocalControllerConfig{
		Name:         name,
		ContainerMap: containerMap,
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
	// Stop container if running (ignore error if there is no running container)
	lc.client.ContainerStop(ctx, container.ID, nil)

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
		Links:        containerConfig.Links,
	}

	// Pull image
	_, err := lc.client.ImagePull(ctx, containerConfig.Image, lc.getPullOptions(containerConfig.Image))
	if err != nil {
		fmt.Printf("Failed to pull image: %v\n", err)
		return "", err
	}

	// Create network if it does not exists
	networkName := "local-iofog-network"
	networks, err := lc.client.NetworkList(ctx, types.NetworkListOptions{})
	networkID := ""
	for i := range networks {
		if networks[i].Name == networkName {
			networkID = networks[i].ID
			break
		}
	}

	if networkID == "" {
		networkResponse, err := lc.client.NetworkCreate(ctx, networkName, types.NetworkCreate{
			Driver:         "bridge",
			CheckDuplicate: true,
		})
		if err != nil {
			fmt.Printf("Failed to create network: %v\n", err)
			return "", err
		}
		networkID = networkResponse.ID
	}

	container, err := lc.client.ContainerCreate(ctx, dockerContainerConfig, hostConfig, nil, containerConfig.ContainerName)
	if err != nil {
		fmt.Printf("Failed to create container: %v\n", err)
		return "", err
	}

	// Connect to network
	err = lc.client.NetworkConnect(ctx, networkID, container.ID, nil)
	if err != nil {
		fmt.Printf("Failed to connect container to network: %v\n", err)
		return "", err
	}

	// Start container
	err = lc.client.ContainerStart(ctx, container.ID, types.ContainerStartOptions{})
	if err != nil {
		fmt.Printf("Failed to start container: %v\n", err)
		return "", err
	}

	return container.ID, err
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
