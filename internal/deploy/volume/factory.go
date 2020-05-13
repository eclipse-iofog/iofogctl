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
	Name      string
	Yaml      []byte
}

type executor struct {
	localExecutor  execute.Executor
	remoteExecutor execute.Executor
	Name           string
	namespace      string
	volume         rsc.Volume
}

func (exe executor) GetName() string {
	return "deploying Volume " + exe.Name
}

func (exe executor) Execute() error {
	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	// Check agents exist
	remoteAgents := make([]*rsc.RemoteAgent, 0)
	localAgents := make([]*rsc.LocalAgent, 0)
	for _, agentName := range exe.volume.Agents {
		baseAgent, err := ns.GetAgent(agentName)
		if err != nil {
			return err
		}
		agent, ok := baseAgent.(*rsc.RemoteAgent)
		if ok {
			// Check SSH details
			if err = agent.ValidateSSH(); err != nil {
				return err
			}
			// Check agent is not local
			if util.IsLocalHost(agent.Host) {
				return util.NewError("Volume deployment is not supported for local Agents")
			}
			remoteAgents = append(remoteAgents, agent)
		} else {
			agent, ok := baseAgent.(*rsc.LocalAgent)
			if ok {
				localAgents = append(localAgents, agent)
			} else {
				return util.NewInternalError("Could not convert Agent type")
			}
		}
	}
	exe.localExecutor = localExecutor{
		agents: localAgents,
		volume: exe.volume,
		ns:     ns,
	}
	exe.remoteExecutor = remoteExecutor{
		agents: remoteAgents,
		volume: exe.volume,
		ns:     ns,
	}
	errs := execute.RunExecutors([]execute.Executor{exe.localExecutor, exe.remoteExecutor}, exe.GetName())
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

func NewExecutor(opt Options) (execute.Executor, error) {
	// Unmarshal file
	var volume rsc.Volume
	if err := yaml.UnmarshalStrict(opt.Yaml, &volume); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return nil, err
	}
	// Check Name
	if opt.Name != "" {
		volume.Name = opt.Name
	}
	if err := util.IsLowerAlphanumeric(volume.Name); err != nil {
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
	return executor{
		Name:      volume.Name,
		namespace: opt.Namespace,
		volume:    volume,
	}, nil
}
