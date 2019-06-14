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

package deploycontroller

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/user"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	dockerClient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type localExecutor struct {
	opt    *Options
	client *dockerClient.Client
}

func newLocalExecutor(opt *Options, client *dockerClient.Client) *localExecutor {
	return &localExecutor{
		opt:    opt,
		client: client,
	}
}

func (exe *localExecutor) deployContainer(image, name string, tcpPorts map[string]nat.Port) error {
	ctx := context.Background()

	portSet := nat.PortSet{}
	portMap := nat.PortMap{}

	for hostPort, containerPort := range tcpPorts {
		portSet[containerPort] = struct{}{}
		portMap[containerPort] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: hostPort,
			},
		}
	}

	containerConfig := &container.Config{
		Image:        image,
		ExposedPorts: portSet,
	}
	hostConfig := &container.HostConfig{
		PortBindings: portMap,
	}

	out, err := exe.client.ImagePull(ctx, image, dockerTypes.ImagePullOptions{})
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, out)

	resp, err := exe.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, name)
	if err != nil {
		return err
	}

	return exe.client.ContainerStart(ctx, resp.ID, dockerTypes.ContainerStartOptions{})
}

func (exe *localExecutor) deployContainers() error {
	// Deploy controller image
	controllerImg, exists := exe.opt.Images["controller"]
	if !exists {
		return util.NewInputError("No controller image specified")
	}
	controllerPortMap := make(map[string]nat.Port)
	controllerPort, err := nat.NewPort("tcp", "51121")
	if err != nil {
		return err
	}
	controllerPortMap["51121"] = controllerPort // 51121:51121/tcp
	err = exe.deployContainer(controllerImg, "iofog-controller", controllerPortMap)
	if err != nil {
		return err
	}

	// Deploy controller image
	connectorImg, exists := exe.opt.Images["connector"]
	if !exists {
		return util.NewInputError("No connector image specified")
	}
	connectorPortMap := make(map[string]nat.Port)
	connectorPort, err := nat.NewPort("tcp", "8080")
	connectorPortMap["53321"] = connectorPort // 53321:8080/tcp
	if err != nil {
		return err
	}
	err = exe.deployContainer(connectorImg, "iofog-connector", connectorPortMap)
	if err != nil {
		return err
	}
}

func (exe *localExecutor) Execute() error {
	currUser, err := user.Current()
	if err != nil {
		return err
	}

	err = exe.deployContainers()
	if err != nil {
		return err
	}

	// Update configuration
	configEntry := config.Controller{
		Name:   exe.opt.Name,
		User:   currUser.Username,
		Host:   "0.0.0.0:51121",
		Images: exe.opt.Images,
	}
	err = config.AddController(exe.opt.Namespace, configEntry)
	if err != nil {
		return err
	}

	fmt.Printf("\nController %s/%s successfully deployed.\n", exe.opt.Namespace, exe.opt.Name)

	return config.Flush()
}
