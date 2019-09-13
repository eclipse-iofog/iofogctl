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

//
// Install iofog-agent stack on an agent host
//
func (exe *remoteExecutor) Execute() (err error) {
	// Get Control Plane
	controlPlane, err := config.GetControlPlane(exe.namespace)
	if err != nil || len(controlPlane.Controllers) == 0 {
		util.PrintError("You must deploy a Controller to a namespace before deploying any Agents")
		return
	}

	// Connect to agent via SSH
	agent := install.NewRemoteAgent(exe.agent.User, exe.agent.Host, exe.agent.Port, exe.agent.KeyFile, exe.agent.Name)

	// Set version
	agent.SetVersion(exe.agent.Version, exe.agent.PackageCloudToken)

	// Try the install
	err = agent.Bootstrap()
	if err != nil {
		return
	}

	// Configure the agent with Controller details
	uuid, err := agent.Configure(&controlPlane.Controllers[0], install.IofogUser(controlPlane.IofogUser))
	if err != nil {
		return
	}

	// Return the Agent through pointer
	exe.agent.UUID = uuid
	exe.agent.Created = util.NowUTC()

	return
}
