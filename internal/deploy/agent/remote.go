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

	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

type remoteExecutor struct {
	namespace string
	agent     *rsc.RemoteAgent
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
	agent, err := install.NewRemoteAgent(exe.agent.SSH.User,
		exe.agent.Host,
		exe.agent.SSH.Port,
		exe.agent.SSH.KeyFile,
		exe.agent.Name,
		exe.agent.UUID)
	if err != nil {
		return "", err
	}

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
			return "", util.NewError("Failed to retrieve Controller endpoint!")
		}
	}

	// Configure the agent with Controller details
	user := install.IofogUser(controlPlane.GetUser())
	user.Password = controlPlane.GetUser().GetRawPassword()
	return agent.Configure(controllerEndpoint, user)
}

// Deploy iofog-agent stack on an agent host
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
	agent, err := install.NewRemoteAgent(exe.agent.SSH.User,
		exe.agent.Host,
		exe.agent.SSH.Port,
		exe.agent.SSH.KeyFile,
		exe.agent.Name,
		exe.agent.UUID)
	if err != nil {
		return err
	}

	// Set custom scripts
	if exe.agent.Scripts != nil {
		if err := agent.CustomizeProcedures(
			exe.agent.Scripts.Directory,
			&exe.agent.Scripts.AgentProcedures); err != nil {
			return err
		}
	}

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
	return nil
}

func ValidateRemoteAgent(agent *rsc.RemoteAgent) error {
	if err := util.IsLowerAlphanumeric("Agent", agent.Name); err != nil {
		return err
	}
	if agent.Name == iofog.VanillaRouterAgentName {
		return util.NewInputError(fmt.Sprintf("%s is a reserved name and cannot be used for an Agent", iofog.VanillaRouterAgentName))
	}
	if (agent.Host != "localhost" && agent.Host != "127.0.0.1") && (agent.Host == "" || agent.SSH.User == "" || agent.SSH.KeyFile == "") {
		return util.NewInputError("For Agents you must specify non-empty values for host, user, and keyfile")
	}
	return nil
}
