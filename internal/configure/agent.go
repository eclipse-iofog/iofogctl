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
	useDetached bool
}

func newAgentExecutor(opt Options) *agentExecutor {
	return &agentExecutor{
		namespace:   opt.Namespace,
		name:        opt.Name,
		keyFile:     opt.KeyFile,
		user:        opt.User,
		port:        opt.Port,
		useDetached: opt.UseDetached,
	}
}

func (exe *agentExecutor) GetName() string {
	return exe.name
}

func (exe *agentExecutor) Execute() error {
	var baseAgent rsc.Agent
	var err error
	if exe.useDetached {
		baseAgent, err = config.GetDetachedAgent(exe.name)
	} else {
		baseAgent, err = config.GetAgent(exe.namespace, exe.name)
	}
	if err != nil {
		return err
	}

	switch agent := baseAgent.(type) {
	case *rsc.LocalAgent:
		return util.NewInputError("Cannot configure Local Agent")
	case *rsc.RemoteAgent:
		// Only updated fields specified
		if exe.user != "" {
			agent.SSH.User = exe.user
		}
		if exe.port != 0 {
			agent.SSH.Port = exe.port
		}
		if exe.keyFile != "" {
			agent.SSH.KeyFile, err = util.FormatPath(exe.keyFile)
			if err != nil {
				return err
			}
		}
		agent.Sanitize()

		// Save config
		if exe.useDetached {
			config.UpdateDetachedAgent(agent)
			return nil
		}
		if err = config.UpdateAgent(exe.namespace, agent); err != nil {
			return err
		}
		return config.Flush()
	}
	return util.NewError("Could not convert Agent to dynamic type")
}
