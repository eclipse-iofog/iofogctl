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

package deployvolume

import (
	"os"

	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
	yaml "gopkg.in/yaml.v2"
)

type Options struct {
	Namespace string
	Yaml      []byte
}

type executor struct {
	localExecutor  execute.Executor
	remoteExecutor execute.Executor
	Name           string
}

func (exe executor) GetName() string {
	return "deploying Volume " + exe.Name
}

func (exe executor) Execute() error {
	errs := execute.RunExecutors([]execute.Executor{exe.localExecutor, exe.remoteExecutor}, exe.GetName())
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	// Unmarshal file
	var volume rsc.Volume
	if err = yaml.UnmarshalStrict(opt.Yaml, &volume); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}
	// Check Name
	if err := util.IsLowerAlphanumeric(volume.Name); err != nil {
		return nil, err
	}
	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return nil, err
	}
	// Check if source is a folder
	info, err := os.Stat(volume.Source)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, util.NewInputError("Source must be a directory")
	}
	// Check agents exist
	remoteAgents := make([]*rsc.RemoteAgent, 0)
	localAgents := make([]*rsc.LocalAgent, 0)
	for _, agentName := range volume.Agents {
		baseAgent, err := ns.GetAgent(agentName)
		if err != nil {
			return nil, err
		}
		agent, ok := baseAgent.(*rsc.RemoteAgent)
		if ok {
			// Check SSH details
			if err = agent.ValidateSSH(); err != nil {
				return nil, err
			}
			// Check agent is not local
			if util.IsLocalHost(agent.Host) {
				return nil, util.NewError("Volume deployment is not supported for local Agents")
			}
			remoteAgents = append(remoteAgents, agent)
		} else {
			agent, ok := baseAgent.(*rsc.LocalAgent)
			if ok {
				localAgents = append(localAgents, agent)
			} else {
				return nil, util.NewInternalError("Could not convert Agent type")
			}
		}
	}
	return executor{
		Name: volume.Name,
		localExecutor: localExecutor{
			agents: localAgents,
			volume: volume,
			ns:     ns,
		},
		remoteExecutor: remoteExecutor{
			agents: remoteAgents,
			volume: volume,
			ns:     ns,
		},
	}, nil
}
