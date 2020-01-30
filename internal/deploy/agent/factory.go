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

	"github.com/eclipse-iofog/iofogctl/internal"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type facadeExecutor struct {
	exe         AgentExecutor
	agent       *config.Agent
	agentConfig *config.AgentConfiguration
	namespace   string
}

type AgentExecutor interface {
	execute.Executor
	SetAgentConfig(*config.AgentConfiguration)
}

func (facade facadeExecutor) SetAgentConfig(config *config.AgentConfiguration) {
	facade.exe.SetAgentConfig(config)
}

func (facade facadeExecutor) Execute() (err error) {
	// Check the namespace exists
	ns, err := config.GetNamespace(facade.namespace)
	if err != nil {
		return err
	}

	// Check Controller exists
	if len(ns.ControlPlane.Controllers) == 0 {
		return util.NewInputError("This namespace does not have a Controller. You must first deploy a Controller before deploying Agents")
	}

	isSystemAgent := facade.agentConfig != nil && internal.IsSystemAgent(*facade.agentConfig)

	util.SpinStart(fmt.Sprintf("Deploying agent %s", facade.GetName()))

	if err = facade.exe.Execute(); err != nil {
		return
	}
	// Don't add system agent to the namespace config file
	if !isSystemAgent {
		if err = config.UpdateAgent(facade.namespace, *facade.agent); err != nil {
			return
		}
	}
	return config.Flush()
}

func (facade facadeExecutor) GetName() string {
	return facade.exe.GetName()
}

func (facade facadeExecutor) ProvisionAgent() (string, error) {
	// Required for attach
	provisionExecutor, ok := facade.exe.(execute.ProvisioningExecutor)
	if !ok {
		return "", util.NewInternalError("Facade executor: Could not convert executor")
	}
	return provisionExecutor.ProvisionAgent()
}

func newFacadeExecutor(exe AgentExecutor, namespace string, agent *config.Agent) execute.Executor {
	return facadeExecutor{
		exe:       exe,
		namespace: namespace,
		agent:     agent,
	}
}

func NewDeployExecutor(namespace string, agent *config.Agent) (execute.Executor, error) {
	if err := util.IsLowerAlphanumeric(agent.Name); err != nil {
		return nil, err
	}

	// Local executor
	if util.IsLocalHost(agent.Host) {
		cli, err := install.NewLocalContainerClient()
		if err != nil {
			return nil, err
		}
		exe, err := newLocalExecutor(namespace, agent, cli)
		if err != nil {
			return nil, err
		}
		return newFacadeExecutor(exe, namespace, agent), nil
	}

	// Default executor
	if agent.Host == "" || agent.SSH.KeyFile == "" || agent.SSH.User == "" {
		return nil, util.NewInputError("Must specify user, host, and key file flags for remote deployment or provisioning")
	}
	return newFacadeExecutor(newRemoteExecutor(namespace, agent), namespace, agent), nil
}
