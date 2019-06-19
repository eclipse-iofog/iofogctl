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

package deployagent

import (
	"fmt"
	"os/user"
	"regexp"

	pb "github.com/schollz/progressbar"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type localExecutor struct {
	opt              *Options
	client           *iofog.LocalContainer
	localAgentConfig *iofog.LocalAgentConfig
	pb               *pb.ProgressBar
}

func getController(namespace string) (*config.Controller, error) {
	controllers, err := config.GetControllers(namespace)
	if err != nil {
		println("You must deploy a Controller to a namespace before deploying any Agents")
		return nil, err
	}
	if len(controllers) != 1 {
		return nil, util.NewInternalError("Only support 1 controller per namespace")
	}
	return &controllers[0], nil
}

func newLocalExecutor(opt *Options, client *iofog.LocalContainer) (*localExecutor, error) {
	// Get controllerConfig
	controller, err := getController(opt.Namespace)
	if err != nil {
		return nil, err
	}
	// Get Controller LocalContainerConfig
	localControllerConfig := iofog.NewLocalControllerConfig(controller.Name, make(map[string]string))
	controllerContainerConfig, _ := localControllerConfig.ContainerMap["controller"]
	return &localExecutor{
		opt:              opt,
		client:           client,
		localAgentConfig: iofog.NewLocalAgentConfig(opt.Name, opt.Image, controllerContainerConfig),
		pb:               pb.New(100),
	}, nil
}

func (exe *localExecutor) provisionAgent() (string, error) {
	// Get agent
	agent := iofog.NewLocalAgent(exe.localAgentConfig, exe.client)

	// Get Controller details
	controller, err := getController(exe.opt.Namespace)
	if err != nil {
		return "", err
	}
	user := iofog.User{
		Name:     controller.IofogUser.Name,
		Surname:  controller.IofogUser.Surname,
		Email:    controller.IofogUser.Email,
		Password: controller.IofogUser.Password,
	}

	// Configure the agent with Controller details
	return agent.Configure(controller, user)
}

func (exe *localExecutor) Execute() error {
	exe.pb.Add(1)
	defer exe.pb.Clear()
	// Get current user
	currUser, err := user.Current()
	if err != nil {
		return err
	}

	// Deploy agent image
	if exe.opt.Image == "" {
		exe.opt.Image = exe.localAgentConfig.DefaultImage
	}

	if _, err = exe.client.DeployContainer(&exe.localAgentConfig.LocalContainerConfig); err != nil {
		return err
	}
	exe.pb.Add(25)

	agentContainerName := exe.localAgentConfig.ContainerName

	// Wait for agent
	if err = exe.client.WaitForCommand(
		regexp.MustCompile("401 Unauthorized"),
		"curl",
		"--request",
		"GET",
		"--url",
		fmt.Sprintf("http://%s:%s/v2/status", exe.localAgentConfig.Host, exe.localAgentConfig.Ports[0].Host),
	); err != nil {
		if cleanErr := exe.client.CleanContainer(agentContainerName); cleanErr != nil {
			fmt.Printf("Could not clean container %s\n", agentContainerName)
		}
		return err
	}
	exe.pb.Add(25)

	// Provision agent
	uuid, err := exe.provisionAgent()
	if err != nil {
		if cleanErr := exe.client.CleanContainer(agentContainerName); cleanErr != nil {
			fmt.Printf("Could not clean container %s\n", agentContainerName)
		}
		return err
	}
	exe.pb.Add(25)

	// Update configuration
	agentIP := fmt.Sprintf("%s:%s", exe.localAgentConfig.Host, exe.localAgentConfig.Ports[0].Host)
	configEntry := config.Agent{
		Name: exe.opt.Name,
		User: currUser.Username,
		Host: agentIP,
		UUID: uuid,
	}
	err = config.AddAgent(exe.opt.Namespace, configEntry)
	if err != nil {
		if cleanErr := exe.client.CleanContainer(agentContainerName); cleanErr != nil {
			fmt.Printf("Could not clean container %s\n", agentContainerName)
		}
		return err
	}

	exe.pb.Add(24)
	if err = config.Flush(); err != nil {
		if cleanErr := exe.client.CleanContainer(agentContainerName); cleanErr != nil {
			fmt.Printf("Could not clean container %s\n", agentContainerName)
		}
		return err
	}

	fmt.Printf("\nAgent %s/%s successfully deployed.\n", exe.opt.Namespace, exe.opt.Name)
	return nil
}
