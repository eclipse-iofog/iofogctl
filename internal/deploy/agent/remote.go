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
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type remoteExecutor struct {
	namespace string
	agent     *config.Agent
	uuid      string
}

func newRemoteExecutor(namespace string, agent *config.Agent) *remoteExecutor {
	exe := &remoteExecutor{}
	exe.namespace = namespace
	exe.agent = agent

	return exe
}

func (exe *remoteExecutor) GetName() string {
	return exe.agent.Name
}

func (exe *remoteExecutor) ProvisionAgent() (string, error) {
	// Get agent
	agent := install.NewRemoteAgent(
		exe.agent.SSH.User,
		exe.agent.Host,
		exe.agent.SSH.Port,
		exe.agent.SSH.KeyFile,
		exe.agent.Name,
		exe.agent.UUID)

	controlPlane, err := config.GetControlPlane(exe.namespace)
	if err != nil {
		return "", err
	}
	controllerEndpoint, err := controlPlane.GetControllerEndpoint()
	if err != nil {
		return "", util.NewError("Failed to retrieve Controller endpoint!")
	}

	// Configure the agent with Controller details
	return agent.Configure(controllerEndpoint, install.IofogUser(controlPlane.IofogUser))
}

//
// Deploy iofog-agent stack on an agent host
//
func (exe *remoteExecutor) Execute() (err error) {
	// Get Control Plane
	controlPlane, err := config.GetControlPlane(exe.namespace)
	if err != nil || len(controlPlane.Controllers) == 0 {
		util.PrintError("You must deploy a Controller to a namespace before deploying any Agents")
		return
	}

	// Connect to agent via SSH
	agent := install.NewRemoteAgent(
		exe.agent.SSH.User,
		exe.agent.Host,
		exe.agent.SSH.Port,
		exe.agent.SSH.KeyFile,
		exe.agent.Name,
		exe.agent.UUID)

	// Set version
	agent.SetVersion(exe.agent.Package.Version)
	agent.SetRepository(exe.agent.Package.Repo, exe.agent.Package.Token)

	// Try the deploy
	err = agent.Bootstrap()
	if err != nil {
		return
	}

	UUID, err := exe.ProvisionAgent()
	if err != nil {
		return err
	}

	// Return the Agent through pointer
	exe.agent.UUID = UUID
	exe.agent.Created = util.NowUTC()

	return
}
