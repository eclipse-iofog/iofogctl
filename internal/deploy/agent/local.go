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
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/client"
	"os/user"
	"regexp"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	install "github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type localExecutor struct {
	namespace        string
	agent            config.Agent
	client           *install.LocalContainer
	localAgentConfig *install.LocalAgentConfig
}

func getController(namespace string) (*config.Controller, error) {
	controllers, err := config.GetControllers(namespace)
	if err != nil {
		fmt.Print("You must deploy a Controller to a namespace before deploying any Agents")
		return nil, err
	}
	if len(controllers) != 1 {
		return nil, util.NewInternalError("Only support 1 controller per namespace")
	}
	return &controllers[0], nil
}

func newLocalExecutor(namespace string, agent config.Agent, client *install.LocalContainer) (*localExecutor, error) {
	// Get controllerConfig
	controller, err := getController(namespace)
	if err != nil {
		return nil, err
	}
	// Get Controller LocalContainerConfig
	localControllerConfig := install.NewLocalControllerConfig(controller.Name, make(map[string]string))
	controllerContainerConfig, _ := localControllerConfig.ContainerMap["controller"]
	return &localExecutor{
		namespace:        namespace,
		agent:            agent,
		client:           client,
		localAgentConfig: install.NewLocalAgentConfig(agent.Name, agent.Image, controllerContainerConfig),
	}, nil
}

func (exe *localExecutor) provisionAgent() (string, error) {
	// Get agent
	agent := install.NewLocalAgent(exe.localAgentConfig, exe.client)

	// Get Controller details
	controller, err := getController(exe.namespace)
	if err != nil {
		return "", err
	}
	user := client.User{
		Name:     controller.IofogUser.Name,
		Surname:  controller.IofogUser.Surname,
		Email:    controller.IofogUser.Email,
		Password: controller.IofogUser.Password,
	}

	// Configure the agent with Controller details
	return agent.Configure(controller, user)
}

func (exe *localExecutor) execute() error {
	defer util.SpinStop()
	// Get current user
	currUser, err := user.Current()
	if err != nil {
		return err
	}

	// Deploy agent image
	util.SpinStart("Deploying Agent container")
	if exe.agent.Image == "" {
		exe.agent.Image = exe.localAgentConfig.DefaultImage
	}

	if _, err = exe.client.DeployContainer(&exe.localAgentConfig.LocalContainerConfig); err != nil {
		return err
	}

	agentContainerName := exe.localAgentConfig.ContainerName

	// Wait for agent
	util.SpinStart("Waiting for Agent")
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

	// Provision agent
	util.SpinStart("Provisioning Agent")
	uuid, err := exe.provisionAgent()
	if err != nil {
		if cleanErr := exe.client.CleanContainer(agentContainerName); cleanErr != nil {
			fmt.Printf("Could not clean container %s\n", agentContainerName)
		}
		return err
	}

	// Update configuration
	agentIP := fmt.Sprintf("%s:%s", exe.localAgentConfig.Host, exe.localAgentConfig.Ports[0].Host)
	configEntry := config.Agent{
		Name: exe.agent.Name,
		User: currUser.Username,
		Host: agentIP,
		UUID: uuid,
	}
	err = config.AddAgent(exe.namespace, configEntry)
	if err != nil {
		if cleanErr := exe.client.CleanContainer(agentContainerName); cleanErr != nil {
			fmt.Printf("Could not clean container %s\n", agentContainerName)
		}
		return err
	}

	if err = config.Flush(); err != nil {
		if cleanErr := exe.client.CleanContainer(agentContainerName); cleanErr != nil {
			fmt.Printf("Could not clean container %s\n", agentContainerName)
		}
		return err
	}

	return nil
}
