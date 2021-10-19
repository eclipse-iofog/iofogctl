/*
 *  *******************************************************************************
 *  * Copyright (c) 2020 Edgeworx, Inc.
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
	"regexp"

	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

type localExecutor struct {
	isSystem         bool
	namespace        string
	agent            *rsc.LocalAgent
	client           *install.LocalContainer
	localAgentConfig *install.LocalAgentConfig
}

func newLocalExecutor(namespace string, agent *rsc.LocalAgent, isSystem bool) (*localExecutor, error) {
	client, err := install.NewLocalContainerClient()
	if err != nil {
		return nil, err
	}
	// Get Controller LocalContainerConfig
	controllerContainerConfig := install.NewLocalControllerConfig("", install.Credentials{})
	return &localExecutor{
		isSystem:  isSystem,
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
			},
			isSystem),
	}, nil
}

func (exe *localExecutor) ProvisionAgent() (string, error) {
	// Get agent
	agent := install.NewLocalAgent(exe.localAgentConfig, exe.client)

	// Get user
	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return "", err
	}
	controlPlane, err := ns.GetControlPlane()
	if err != nil {
		return "", err
	}
	// Try Agent-specific endpoint first
	controllerEndpoint := exe.agent.GetControllerEndpoint()
	if controllerEndpoint == "" {
		controllerEndpoint, err = controlPlane.GetEndpoint()
		if err != nil {
			return "", err
		}
	}

	// Configure the agent with Controller details
	user := install.IofogUser(controlPlane.GetUser())
	user.Password = controlPlane.GetUser().GetRawPassword()
	return agent.Configure(controllerEndpoint, user)
}

func (exe *localExecutor) GetName() string {
	return exe.agent.Name
}

func (exe *localExecutor) Execute() error {
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

	if _, err := exe.client.DeployContainer(&exe.localAgentConfig.LocalContainerConfig); err != nil {
		return err
	}

	// Wait for agent
	util.SpinStart("Waiting for Agent")
	if err := exe.client.WaitForCommand(
		install.GetLocalContainerName("agent", exe.isSystem),
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
	exe.agent.Host = fmt.Sprintf("%s:%s", exe.localAgentConfig.Host, exe.localAgentConfig.Ports[0].Host)
	exe.agent.UUID = uuid

	return nil
}
