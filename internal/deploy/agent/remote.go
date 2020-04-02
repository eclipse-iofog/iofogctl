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

	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type remoteExecutor struct {
	namespace string
	agent     *rsc.RemoteAgent
	uuid      string
}

func newRemoteExecutor(namespace string, agent *rsc.RemoteAgent) *remoteExecutor {
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

	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return "", err
	}
	controlPlane, err := ns.GetControlPlane()
	if err != nil {
		return "", err
	}
	controllerEndpoint, err := controlPlane.GetEndpoint()
	if err != nil {
		return "", util.NewError("Failed to retrieve Controller endpoint!")
	}

	// Configure the agent with Controller details
	return agent.Configure(controllerEndpoint, install.IofogUser(controlPlane.GetUser()))
}

//
// Deploy iofog-agent stack on an agent host
//
func (exe *remoteExecutor) Execute() (err error) {
	// Get Control Plane
	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	controlPlane, err := ns.GetControlPlane()
	if err != nil || len(controlPlane.GetControllers()) == 0 {
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

	uuid, err := exe.ProvisionAgent()
	if err != nil {
		return err
	}

	// Return the Agent through pointer
	exe.agent.UUID = uuid
	exe.agent.Created = util.NowUTC()
	return
}

func ValidateRemoteAgent(agent rsc.RemoteAgent) error {
	if agent.Name == "" {
		return util.NewInputError("You must specify a non-empty value for name value of Agents")
	}
	if agent.Name == iofog.VanillaRouterAgentName {
		return util.NewInputError(fmt.Sprintf("%s is a reserved name and cannot be used for an Agent", iofog.VanillaRouterAgentName))
	}
	if (agent.Host != "localhost" && agent.Host != "127.0.0.1") && (agent.Host == "" || agent.SSH.User == "" || agent.SSH.KeyFile == "") {
		return util.NewInputError("For Agents you must specify non-empty values for host, user, and keyfile")
	}
	return nil
}
