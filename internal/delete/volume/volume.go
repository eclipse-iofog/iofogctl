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

package deletevolume

import (
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type Executor struct {
	namespace string
	volume    rsc.Volume
}

func NewExecutor(namespace, name string) (execute.Executor, error) {
	volume, err := config.GetVolume(namespace, name)
	if err != nil {
		return nil, err
	}
	exe := &Executor{
		namespace: namespace,
		volume:    volume,
	}

	return exe, nil
}

// GetName returns application name
func (exe *Executor) GetName() string {
	return exe.volume.Name
}

// Execute deletes application by deleting its associated flow
func (exe *Executor) Execute() error {
	util.SpinStart("Deleting Volume")

	// Delete files
	ch := make(chan error, len(exe.volume.Agents))
	for idx := range exe.volume.Agents {
		go exe.execute(idx, ch)
	}
	for idx := 0; idx < len(exe.volume.Agents); idx++ {
		if err := <-ch; err != nil {
			return err
		}
	}

	// Delete from config
	if err := config.DeleteVolume(exe.namespace, exe.volume.Name); err != nil {
		return err
	}
	return config.Flush()
}

func (exe *Executor) execute(agentIdx int, ch chan error) {
	agentName := exe.volume.Agents[agentIdx]
	baseAgent, err := config.GetAgent(exe.namespace, agentName)
	if err != nil {
		ch <- err
	}
	agent, ok := baseAgent.(*rsc.RemoteAgent)
	if !ok {
		ch <- util.NewInputError("Cannot delete Volumes from Local Agents")
	}
	// Check SSH details
	if err := agent.ValidateSSH(); err != nil {
		ch <- err
	}
	// Check agent is not local
	if util.IsLocalHost(agent.Host) {
		ch <- util.NewError("Volume deletion is not supported for local Agents")
	}
	// Connect
	ssh := util.NewSecureShellClient(agent.SSH.User, agent.Host, agent.SSH.KeyFile)
	if err := ssh.Connect(); err != nil {
		ch <- err
	}
	// Delete
	if _, err := ssh.Run("rm -rf " + util.AddTrailingSlash(exe.volume.Destination) + "*"); err != nil {
		ch <- err
	}

	ch <- nil
}
