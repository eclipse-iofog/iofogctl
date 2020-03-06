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

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"gopkg.in/yaml.v2"
)

type Options struct {
	Namespace string
	Yaml      []byte
}

type remoteExecutor struct {
	volume    config.Volume
	namespace string
	agents    []config.Agent
}

func (exe remoteExecutor) GetName() string {
	return "Deploy Volume " + exe.volume.Destination
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
	return nil
}

func (exe remoteExecutor) execute(agentIdx int, ch chan error) {
	agent := exe.agents[agentIdx]

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
	var volume config.Volume
	if err = yaml.UnmarshalStrict(opt.Yaml, &volume); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}
	// Check agents exist
	agents := make([]config.Agent, 0)
	for _, agentName := range volume.Agents {
		agent, err := config.GetAgent(opt.Namespace, agentName)
		if err != nil {
			return nil, err
		}
		// Check SSH details
		if agent.Host == "" || agent.SSH.User == "" || agent.SSH.Port == 0 || agent.SSH.KeyFile == "" {
			return nil, util.NewInputError("Trying to push volume but SSH details for Agent " + agent.Name + " are not available. They can be added manually through the `configure` command")
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
