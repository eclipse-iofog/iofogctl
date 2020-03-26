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

package deployvolume

import (
	"errors"
	"fmt"

	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
	"gopkg.in/yaml.v2"
)

type Options struct {
	Namespace string
	Yaml      []byte
}

type remoteExecutor struct {
	volume    rsc.Volume
	namespace string
	agents    []rsc.Agent
}

func (exe remoteExecutor) GetName() string {
	return "deploying Volume " + exe.volume.Name
}

func (exe remoteExecutor) Execute() error {
	util.SpinStart("Pushing volumes to Agents")
	// Transfer files
	ch := make(chan error, len(exe.volume.Agents))
	for idx := range exe.volume.Agents {
		go exe.execute(idx, ch)
	}
	for idx := 0; idx < len(exe.volume.Agents); idx++ {
		if err := <-ch; err != nil {
			return err
		}
	}
	// Update config
	if err := config.AddVolume(exe.namespace, exe.volume); err != nil {
		return err
	}
	return config.Flush()
}

func (exe remoteExecutor) execute(agentIdx int, ch chan error) {
	agent := exe.agents[agentIdx].(*rsc.RemoteAgent)

	// Connect
	ssh := util.NewSecureShellClient(agent.SSH.User, agent.Host, agent.SSH.KeyFile)
	if err := ssh.Connect(); err != nil {
		msg := `Failed to Connect to Agent %s.
%s`
		ch <- errors.New(fmt.Sprintf(msg, agent.Name, err.Error()))
		return
	}
	defer ssh.Disconnect()

	// Create base path
	if err := ssh.CreateFolder(exe.volume.Destination); err != nil {
		msg := `Failed to create base directory %s on Agent %s.
%s`
		ch <- errors.New(fmt.Sprintf(msg, exe.volume.Destination, agent.Name, err.Error()))
		return
	}
	// Copy volume
	if err := ssh.CopyFolderTo(exe.volume.Source, exe.volume.Destination, exe.volume.Permissions, true); err != nil {
		msg := `Failed to copy volume to Agent %s.
%s`
		ch <- errors.New(fmt.Sprintf(msg, agent.Name, err.Error()))
		return
	}

	ch <- nil
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	// Unmarshal file
	var volume rsc.Volume
	if err = yaml.UnmarshalStrict(opt.Yaml, &volume); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}
	// Check agents exist
	agents := make([]rsc.Agent, 0)
	for _, agentName := range volume.Agents {
		baseAgent, err := config.GetAgent(opt.Namespace, agentName)
		if err != nil {
			return nil, err
		}
		agent, ok := baseAgent.(*rsc.RemoteAgent)
		if !ok {
			return nil, util.NewInputError("Cannot push Volumes to Local Agents")
		}
		// Check SSH details
		if agent.Host == "" || agent.SSH.User == "" || agent.SSH.Port == 0 || agent.SSH.KeyFile == "" {
			return nil, util.NewInputError("Trying to push volume but SSH details for Agent " + agent.Name + " are not available. They can be added manually through the `configure` command")
		}
		// Check agent is not local
		if util.IsLocalHost(agent.Host) {
			return nil, util.NewError("Volume deployment is not supported for local Agents")
		}
		// Record all agent details
		agents = append(agents, agent)
	}
	return remoteExecutor{
		agents:    agents,
		volume:    volume,
		namespace: opt.Namespace,
	}, nil
}
