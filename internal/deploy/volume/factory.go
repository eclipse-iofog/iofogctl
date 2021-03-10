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

	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	"github.com/eclipse-iofog/iofogctl/v3/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
	yaml "gopkg.in/yaml.v2"
)

type Options struct {
	Namespace string
	Name      string
	Yaml      []byte
}

type executor struct {
	Name      string
	namespace string
	volume    rsc.Volume
}

func (exe *executor) GetName() string {
	return "deploying Volume " + exe.Name
}

func (exe *executor) Execute() error {
	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	// Check agents exist
	remoteAgents := []*rsc.RemoteAgent{}
	localAgents := []*rsc.LocalAgent{}
	for _, agentName := range exe.volume.Agents {
		baseAgent, err := ns.GetAgent(agentName)
		if err != nil {
			return err
		}
		agent, ok := baseAgent.(*rsc.RemoteAgent)
		if ok {
			// Check SSH details
			if err := agent.ValidateSSH(); err != nil {
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
	executors := []execute.Executor{}
	if len(localAgents) > 0 {
		executors = append(executors, &localExecutor{
			agents: localAgents,
			volume: exe.volume,
			ns:     ns,
		})
	}
	if len(remoteAgents) > 0 {
		executors = append(executors, &remoteExecutor{
			agents: remoteAgents,
			volume: exe.volume,
			ns:     ns,
		})
	}
	if errs := execute.RunExecutors(executors, exe.GetName()); len(errs) > 0 {
		return execute.CoalesceErrors(errs)
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
	if err := util.IsLowerAlphanumeric("Volume", volume.Name); err != nil {
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
	return &executor{
		Name:      volume.Name,
		namespace: opt.Namespace,
		volume:    volume,
	}, nil
}
