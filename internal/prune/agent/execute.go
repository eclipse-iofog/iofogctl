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

package pruneagent

import (
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
}

func NewExecutor(namespace, name string, useDetached bool) execute.Executor {
	return executor{name: name, namespace: namespace, useDetached: useDetached}
}

func (exe executor) GetName() string {
	return exe.name
}

func (exe executor) Execute() error {
	util.SpinStart("Pruning Agent")
	// Update local cache based on Controller
	err := iutil.UpdateAgentCache(exe.namespace)
	if err != nil {
		return err
	}

	var baseAgent rsc.Agent
	if exe.useDetached {
		baseAgent, err = config.GetDetachedAgent(exe.name)
	} else {
		baseAgent, err = config.GetAgent(exe.namespace, exe.name)
	}
	if err != nil {
		return err
	}
	// Prune Agent
	switch agent := baseAgent.(type) {
	case *rsc.LocalAgent:
		if err = exe.localAgentPrune(); err != nil {
			return err
		}
	case *rsc.RemoteAgent:
		if exe.useDetached {
			if err = exe.remoteDetachedAgentPrune(agent); err != nil {
				return err
			}
		} else {
			if err = exe.remoteAgentPrune(baseAgent); err != nil {
				return err
			}
		}
	}
	return nil
}
