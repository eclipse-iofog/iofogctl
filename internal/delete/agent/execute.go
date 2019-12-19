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

package deleteagent

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type executor struct {
	name      string
	namespace string
}

func NewExecutor(namespace, name string) (execute.Executor, error) {
	return executor{name: name, namespace: namespace}, nil
}

func (exe executor) GetName() string {
	return exe.name
}

func (exe executor) Execute() error {
	util.SpinStart("Deleting Agent")

	// Delete agent software first, so it can properly deprovision itself before being removed
	// Get Agent from config
	agent, err := config.GetAgent(exe.namespace, exe.name)
	if err == nil {
		if util.IsLocalHost(agent.Host) {
			if err = exe.deleteLocalContainer(); err != nil {
				util.PrintInfo(fmt.Sprintf("Could not remove iofog-agent container %s. Error: %s\n", agent.Host, err.Error()))
			}
		} else {
			if err = exe.deleteRemoteAgent(agent); err != nil {
				util.PrintInfo(fmt.Sprintf("Could not remove iofog-agent from the remote host %s. Error: %s\n", agent.Host, err.Error()))
			}
		}
		if err = config.DeleteAgent(exe.namespace, exe.name); err != nil {
			util.PrintInfo(fmt.Sprintf("Could not remove iofog-agent from iofogctl config. Error: %s\n", err.Error()))
		} else {
			defer config.Flush()
		}
	} else {
		util.PrintInfo(fmt.Sprintf("Could not find iofog-agent in iofogctl config. Error: %s\n", err.Error()))
	}

	// Try to get a Controller client to talk to the REST API
	ctrl, err := internal.NewControllerClient(exe.namespace)
	if err == nil {
		// Does agent exists on Controller
		agent, err := ctrl.GetAgentByName(exe.name)
		if err != nil {
			util.PrintInfo(fmt.Sprintf("Could not delete agent %s from the Controller. Error: %s\n", exe.name, err.Error()))
		} else {
			// Perform deletion of Agent through Controller
			if err = ctrl.DeleteAgent(agent.UUID); err != nil {
				util.PrintInfo(fmt.Sprintf("Could not delete agent %s from the Controller. Error: %s\n", exe.name, err.Error()))
			}
		}
	} else {
		util.PrintInfo(fmt.Sprintf("Could not delete agent %s from the Controller. Error: %s\n", exe.name, err.Error()))
	}

	return nil
}

func (exe executor) deleteRemoteAgent(agent config.Agent) (err error) {
	// Stop and remove the Agent process on remote server
	if agent.Host == "" || agent.SSH.User == "" || agent.SSH.KeyFile == "" || agent.SSH.Port == 0 {
		util.PrintNotify("Could not stop daemon for Agent " + agent.Name + ". SSH details missing from local cofiguration. Use configure command to add SSH details.")
	} else {
		sshAgent := install.NewRemoteAgent(agent.SSH.User, agent.Host, agent.SSH.Port, agent.SSH.KeyFile, agent.Name)
		if err = sshAgent.Uninstall(); err != nil {
			util.PrintNotify(fmt.Sprintf("Failed to stop daemon on Agent %s. %s", agent.Name, err.Error()))
		}
	}
	return
}
