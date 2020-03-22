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
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type agentExecutor struct {
	namespace   string
	name        string
	keyFile     string
	user        string
	port        int
	host        string
	useDetached bool
}

func newAgentExecutor(opt Options) *agentExecutor {
	return &agentExecutor{
		namespace:   opt.Namespace,
		name:        opt.Name,
		keyFile:     opt.KeyFile,
		user:        opt.User,
		port:        opt.Port,
		host:        opt.Host,
		useDetached: opt.UseDetached,
	}
}

func (exe *agentExecutor) GetName() string {
	return exe.name
}

func (exe *agentExecutor) Execute() error {
	var agent rsc.Agent
	var err error
	if exe.useDetached {
		agent, err = config.GetDetachedAgent(exe.name)
	} else {
		agent, err = config.GetAgent(exe.namespace, exe.name)
	}

	if err != nil {
		return err
	}

	// Only updated fields specified
	if exe.keyFile != "" {
		agent.SSH.KeyFile, err = util.FormatPath(exe.keyFile)
		if err != nil {
			return err
		}
	}
	if exe.host != "" {
		agent.Host = exe.host
	}
	if exe.user != "" {
		agent.SSH.User = exe.user
	}
	if exe.port != 0 {
		agent.SSH.Port = exe.port
	}

	// Add port if not specified or existing
	if agent.SSH.Port == 0 {
		agent.SSH.Port = 22
	}

	// Save config
	if exe.useDetached {
		return config.UpdateDetachedAgent(agent)
	}
	if err = config.UpdateAgent(exe.namespace, agent); err != nil {
		return err
	}

	return config.Flush()
}
