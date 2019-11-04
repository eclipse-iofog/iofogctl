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

package configure

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
)

type agentExecutor struct {
	namespace string
	name      string
	keyFile   string
	user      string
	port      int
}

func newAgentExecutor(opt Options) *agentExecutor {
	return &agentExecutor{
		namespace: opt.Namespace,
		name:      opt.Name,
		keyFile:   opt.KeyFile,
		user:      opt.User,
		port:      opt.Port,
	}
}

func (exe *agentExecutor) GetName() string {
	return exe.name
}

func (exe *agentExecutor) Execute() error {
	// Get config
	agent, err := config.GetAgent(exe.namespace, exe.name)
	if err != nil {
		return err
	}

	if exe.keyFile != "" {
		agent.KeyFile = exe.keyFile
	}

	if exe.user != "" {
		agent.User = exe.user
	}

	if exe.port != 0 {
		agent.Port = exe.port
	}

	// Save config
	if err = config.UpdateAgent(exe.namespace, agent); err != nil {
		return err
	}

	return config.Flush()
}
