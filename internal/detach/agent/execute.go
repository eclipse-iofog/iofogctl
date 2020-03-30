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

package detachagent

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/v2/internal"
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
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
	util.SpinStart("Detaching Agent")

	// Check doesn't already exist with same name
	if _, err := config.GetDetachedAgent(exe.name); err == nil {
		msg := `An Agent with the name '%s' is already detached. Rename one of the Agents and try to detach again:
iofogctl rename agent %s %s-2 -n %s
iofogctl rename agent %s %s-2 -n %s --detached`
		return util.NewConflictError(fmt.Sprintf(msg, exe.name, exe.name, exe.name, exe.namespace, exe.name, exe.name, exe.namespace))
	}

	baseAgent, err := config.GetAgent(exe.namespace, exe.name)
	if err == nil {
		// Deprovision agent
		switch agent := baseAgent.(type) {
		case *rsc.LocalAgent:
			if err = exe.localDeprovision(); err != nil {
				return err
			}
		case *rsc.RemoteAgent:
			if err = exe.remoteDeprovision(agent); err != nil {
				return err
			}
		}
	}

	// Try to get a Controller client to talk to the REST API
	ctrl, err := internal.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	agentInfo, err := ctrl.GetAgentByName(exe.name, false)
	if err != nil {
		return err
	}

	// Perform deletion of Agent through Controller
	if err = ctrl.DeleteAgent(agentInfo.UUID); err != nil {
		return err
	}

	// Try to detach from config
	if err = config.DetachAgent(exe.namespace, exe.name); err != nil {
		return err
	}

	return config.Flush()
}
