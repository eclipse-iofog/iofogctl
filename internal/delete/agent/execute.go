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
	force       bool
}

func NewExecutor(namespace, name string, useDetached, force bool) (execute.Executor, error) {
	return executor{name: name, namespace: namespace, useDetached: useDetached, force: force}, nil
}

func (exe executor) GetName() string {
	return exe.name
}

func (exe executor) Execute() (err error) {
	util.SpinStart("Deleting Agent")

	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}

	// Update Agent cache
	if !exe.useDetached {
		if err := iutil.UpdateAgentCache(exe.namespace); err != nil {
			return err
		}
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
		baseAgent, err = ns.GetAgent(exe.name)
		if err != nil {
			return err
		}
	}

	// Check if it has microservices running on it
	if !exe.useDetached && !exe.force {
		// Try to get a Controller client to talk to the REST API
		ctrl, err := iutil.NewControllerClient(exe.namespace)
		if err != nil {
			return err
		}
		msvcList, err := ctrl.GetAllMicroservices()
		if err != nil {
			return err
		}
		for _, msvc := range msvcList.Microservices {
			if msvc.AgentUUID == baseAgent.GetUUID() {
				return util.NewInputError(fmt.Sprintf("Could not delete Agent %s because it still has microservices running. Remove the microservices first, or use the --force option.", baseAgent.GetName()))
			}
		}
	}

	switch agent := baseAgent.(type) {
	case *rsc.LocalAgent:
		if err = exe.deleteLocalContainer(); err != nil {
			util.PrintInfo(fmt.Sprintf("Could not remove Agent container %s. Error: %s\n", agent.GetHost(), err.Error()))
		}
	case *rsc.RemoteAgent:
		if err = exe.deleteRemoteAgent(agent); err != nil {
			util.PrintInfo(fmt.Sprintf("Could not remove Agent from the remote host %s. Error: %s\n", agent.GetHost(), err.Error()))
		}
	}

	// Remove from Controller
	if !exe.useDetached {
		// Try to get a Controller client to talk to the REST API
		ctrl, err := iutil.NewControllerClient(exe.namespace)
		if err != nil {
			util.PrintInfo(fmt.Sprintf("Could not delete Agent %s from the Controller. Error: %s\n", exe.name, err.Error()))
		}
		// Perform deletion of Agent through Controller
		if err = ctrl.DeleteAgent(baseAgent.GetUUID()); err != nil {
			return err
		}
		if err = ns.DeleteAgent(baseAgent.GetName()); err != nil {
			return err
		}
	} else {
		// Update config
		if err = config.DeleteDetachedAgent(baseAgent.GetName()); err != nil {
			return err
		}
	}

	return config.Flush()
}
