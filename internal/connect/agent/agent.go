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

	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/internal/execute"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type executor struct {
	agent     rsc.Agent
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
			agent.SSH.KeyFile = exe.agent.SSH.KeyFile
			agent.SSH.Port = exe.agent.SSH.Port
			agent.SSH.User = exe.agent.SSH.User
			config.UpdateAgent(exe.namespace, agent)
			return nil
		}
	}

	util.PrintNotify(fmt.Sprintf("ECN does not contain agent %s\n", exe.agent.Name))
	return nil
}

func NewExecutor(namespace, name string, yaml []byte) (execute.Executor, error) {
	// Read the input file
	agent, err := unmarshallYAML(yaml)
	if err != nil {
		return nil, err
	}
	agent.Name = name

	return executor{namespace: namespace, agent: agent}, nil
}
