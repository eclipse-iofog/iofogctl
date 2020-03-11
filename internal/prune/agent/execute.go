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

package pruneagent

import (
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/internal/execute"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type executor struct {
	name        string
	namespace   string
	useDetached bool
}

func NewExecutor(namespace, name string, useDetached bool) execute.Executor {
	return executor{name: name, namespace: namespace, useDetached: useDetached}
}

func (exe executor) GetName() string {
	return exe.name
}

func (exe executor) Execute() error {
	util.SpinStart("Pruning Agent")
	var agent config.Agent
	var err error
	if exe.useDetached {
		agent, err = config.GetDetachedAgent(exe.name)
	} else {
		agent, err = config.GetAgent(exe.namespace, exe.name)
	}
	if err != nil {
		return err
	}
	// Prune Agent
	if util.IsLocalHost(agent.Host) {
		if err = exe.localAgentPrune(); err != nil {
			return err
		}
	} else {
		if exe.useDetached {
			if err = exe.remoteDetachedAgentPrune(agent); err != nil {
				return err
			}
		} else {
			if err = exe.remoteAgentPrune(agent); err != nil {
				return err
			}
		}
	}
	return nil
}
