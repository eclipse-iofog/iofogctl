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

package deleteagent

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	iutil "github.com/eclipse-iofog/iofogctl/v2/internal/util"
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

func (exe executor) Execute() (err error) {
	util.SpinStart("Deleting Agent")

	// Update Agent cache
	if err := iutil.UpdateAgentCache(exe.namespace); err != nil {
		return err
	}

	// Delete agent software first, so it can properly deprovision itself before being removed
	// Get Agent from config
	var baseAgent rsc.Agent
	if exe.useDetached {
		baseAgent, err = config.GetDetachedAgent(exe.name)
		if err != nil {
			return err
		}
	} else {
		baseAgent, err = config.GetAgent(exe.namespace, exe.name)
		if err != nil {
			return err
		}
	}
	if !exe.soft {
		switch agent := baseAgent.(type) {
		case *rsc.LocalAgent:
			if err = exe.deleteLocalContainer(); err != nil {
				util.PrintInfo(fmt.Sprintf("Could not remove iofog-agent container %s. Error: %s\n", agent.GetHost(), err.Error()))
			}
		case *rsc.RemoteAgent:
			if err = exe.deleteRemoteAgent(agent); err != nil {
				util.PrintInfo(fmt.Sprintf("Could not remove iofog-agent from the remote host %s. Error: %s\n", agent.GetHost(), err.Error()))
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

	// Remove from Controller
	if !exe.useDetached {
		// Try to get a Controller client to talk to the REST API
		ctrl, err := iutil.NewControllerClient(exe.namespace)
		if err != nil {
			util.PrintInfo(fmt.Sprintf("Could not delete agent %s from the Controller. Error: %s\n", exe.name, err.Error()))
		}
		// Perform deletion of Agent through Controller
		if err = ctrl.DeleteAgent(baseAgent.GetUUID()); err != nil {
			return err
		}
	}

	return
}
