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

	"github.com/eclipse-iofog/iofogctl/v2/internal"
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/internal/execute"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type executor struct {
	name        string
	namespace   string
	useDetached bool
	soft        bool
}

func NewExecutor(namespace, name string, useDetached, soft bool) (execute.Executor, error) {
	return executor{name: name, namespace: namespace, useDetached: useDetached, soft: soft}, nil
}

func (exe executor) GetName() string {
	return exe.name
}

func (exe executor) Execute() error {
	util.SpinStart("Deleting Agent")

	// Delete agent software first, so it can properly deprovision itself before being removed
	// Get Agent from config
	var agent rsc.Agent
	var err error
	if exe.useDetached {
		agent, err = config.GetDetachedAgent(exe.name)
	} else {
		agent, err = config.GetAgent(exe.namespace, exe.name)
	}
	if err == nil {
		if !exe.soft {
			if util.IsLocalHost(agent.Host) {
				if err = exe.deleteLocalContainer(); err != nil {
					util.PrintInfo(fmt.Sprintf("Could not remove iofog-agent container %s. Error: %s\n", agent.Host, err.Error()))
				}
			} else {
				if err = exe.deleteRemoteAgent(agent); err != nil {
					util.PrintInfo(fmt.Sprintf("Could not remove iofog-agent from the remote host %s. Error: %s\n", agent.Host, err.Error()))
				}
			}
		}
		if exe.useDetached {
			return config.DeleteDetachedAgent(exe.name)
		}
		if err = config.DeleteAgent(exe.namespace, exe.name); err != nil {
			util.PrintInfo(fmt.Sprintf("Could not remove iofog-agent from iofogctl config. Error: %s\n", err.Error()))
		} else {
			defer config.Flush()
		}
	} else {
		return util.NewError(fmt.Sprintf("Could not find Agent in iofogctl config. Please run `iofogctl -n %s get agents` to update your config. Error: %s\n", exe.namespace, err.Error()))
	}

	if !exe.useDetached {
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
	}

	return nil
}
