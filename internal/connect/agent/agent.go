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

package connectagent

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	deployagent "github.com/eclipse-iofog/iofogctl/internal/deploy/agent"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type executor struct {
	agent     config.Agent
	namespace string
}

func (exe executor) GetName() string {
	return exe.agent.Name
}

func (exe executor) Execute() error {
	agents, err := config.GetAgents(exe.namespace)
	if err != nil {
		return err
	}

	for _, agent := range agents {
		if agent.Name == exe.agent.Name {
			// Only update ssh info
			agent.KeyFile = exe.agent.KeyFile
			agent.Port = exe.agent.Port
			agent.User = exe.agent.User
			config.UpdateAgent(exe.namespace, agent)
			return nil
		}
	}

	util.PrintNotify(fmt.Sprintf("ECN does not contain agent %s\n", exe.agent.Name))
	return nil
}

func NewExecutor(name, namespace string, yaml []byte) (execute.Executor, error) {
	// Read the input file
	agent, err := deployagent.UnmarshallYAML(yaml)
	if err != nil {
		return nil, err
	}
	agent.Name = name

	return executor{namespace: namespace, agent: agent}, nil
}
