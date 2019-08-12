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
	agent     config.Agent
	uuid      string
}

func newRemoteExecutor(namespace string, agent config.Agent) *remoteExecutor {
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
func (exe *remoteExecutor) Execute() error {
	defer util.SpinStop()
	util.SpinStart("Deploying agent " + exe.agent.Name)

	configEntry, err := exe.deployAgent()
	if err != nil {
		return err
	}

	if err = config.UpdateAgent(exe.namespace, configEntry); err != nil {
		return err
	}

	return config.Flush()
}

func (exe *remoteExecutor) deployAgent() (configEntry config.Agent, err error) {
	// Get Controllers from namespace
	controllers, err := config.GetControllers(exe.namespace)

	// Do we actually have any controllers?
	if err != nil {
		util.PrintError("You must deploy a Controller to a namespace before deploying any Agents")
		return
	}

	// Connect to agent via SSH
	agent := install.NewRemoteAgent(exe.agent.User, exe.agent.Host, exe.agent.Port, exe.agent.KeyFile, exe.agent.Name)

	// Try the install
	err = agent.Bootstrap()
	if err != nil {
		return
	}

	// Configure the agent with Controller details
	uuid, err := agent.Configure(&controllers[0], install.IofogUser(controllers[0].IofogUser))
	if err != nil {
		return
	}

	configEntry = config.Agent{
		Name:    exe.agent.Name,
		User:    exe.agent.User,
		Host:    exe.agent.Host,
		Port:    exe.agent.Port,
		KeyFile: exe.agent.KeyFile,
		UUID:    uuid,
		Created: util.NowUTC(),
	}
	return
}
