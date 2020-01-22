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

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type localExecutor struct {
	namespace        string
	agent            *config.Agent
	agentConfig      *config.AgentConfiguration
	client           *install.LocalContainer
	localAgentConfig *install.LocalAgentConfig
}

func (exe *localExecutor) SetAgentConfig(config *config.AgentConfiguration) {
	exe.agentConfig = config
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

func newLocalExecutor(namespace string, agent *config.Agent, client *install.LocalContainer) (*localExecutor, error) {
	// Get Controller LocalContainerConfig
	controllerContainerConfig := install.NewLocalControllerConfig("", install.Credentials{})
	return &localExecutor{
		namespace: namespace,
		agent:     agent,
		client:    client,
		localAgentConfig: install.NewLocalAgentConfig(
			agent.Name,
			agent.Container.Image,
			controllerContainerConfig,
			install.Credentials{
				User:     agent.Container.Credentials.User,
				Password: agent.Container.Credentials.Password,
			}),
	}, nil
}

func (exe *localExecutor) ProvisionAgent() (string, error) {
	// Get agent
	agent := install.NewLocalAgent(exe.agentConfig, exe.localAgentConfig, exe.client)

	// Get Controller details
	controller, err := getController(exe.namespace)
	if err != nil {
		return "", err
	}

	// Get user
	controlPlane, err := config.GetControlPlane(exe.namespace)
	if err != nil {
		return "", err
	}

	// Configure the agent with Controller details
	return agent.Configure(controller, install.IofogUser(controlPlane.IofogUser))
}

func (exe *localExecutor) GetName() string {
	return exe.agent.Name
}

func (exe *localExecutor) Execute() error {

	// Get current user
	currUser, err := user.Current()
	if err != nil {
		return err
	}

	// Deploy agent image
	util.SpinStart("Deploying Agent container")
	if exe.agent.Container.Image == "" {
		exe.agent.Container.Image = exe.localAgentConfig.DefaultImage
	}

	// If container already exists, clean it
	agentContainerName := exe.localAgentConfig.ContainerName
	if _, err := exe.client.GetContainerByName(agentContainerName); err == nil {
		if err := exe.client.CleanContainer(agentContainerName); err != nil {
			return err
		}
	}

	if _, err = exe.client.DeployContainer(&exe.localAgentConfig.LocalContainerConfig); err != nil {
		return err
	}

	// Wait for agent
	util.SpinStart("Waiting for Agent")
	if err = exe.client.WaitForCommand(
		install.GetLocalContainerName("agent"),
		regexp.MustCompile("ioFog daemon[ |\t]*: RUNNING"),
		"iofog-agent",
		"status",
	); err != nil {
		if cleanErr := exe.client.CleanContainer(agentContainerName); cleanErr != nil {
			util.PrintNotify(fmt.Sprintf("Could not clean container: %v", agentContainerName))
		}
		return err
	}

	// Provision agent
	util.SpinStart("Provisioning Agent")
	uuid, err := exe.ProvisionAgent()
	if err != nil {
		if cleanErr := exe.client.CleanContainer(agentContainerName); cleanErr != nil {
			util.PrintNotify(fmt.Sprintf("Could not clean container: %v", agentContainerName))
		}
		return err
	}

	// Return new Agent config because variable is a pointer
	exe.agent.SSH.User = currUser.Username
	exe.agent.Host = fmt.Sprintf("%s:%s", exe.localAgentConfig.Host, exe.localAgentConfig.Ports[0].Host)
	exe.agent.UUID = uuid

	return nil
}
