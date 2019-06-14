package iofog

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// LocalContainer struct to encapsulate utilities around docker
type LocalContainer struct {
	client *client.Client
}

type LocalContainerPort struct {
	Protocol string
	Port     string
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

// DeployContainer deploys a container based on an image and a port mappin
func (lc *LocalContainer) DeployContainer(image, name string, ports map[string]*LocalContainerPort) error {
	ctx := context.Background()

	portSet := nat.PortSet{}
	portMap := nat.PortMap{}

	for hostPort, containerPort := range ports {
		natPort, err := nat.NewPort(containerPort.Protocol, containerPort.Port)
		if err != nil {
			return err
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
		return err
	}
	io.Copy(os.Stdout, out)

	resp, err := lc.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, name)
	if err != nil {
		return err
	}

	return lc.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
}
