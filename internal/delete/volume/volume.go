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
	ns         *rsc.Namespace
	volumeName string
}

func NewExecutor(namespace, name string) (execute.Executor, error) {
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return nil, err
	}
	exe := &Executor{
		ns:         ns,
		volumeName: name,
	}

	return exe, nil
}

// GetName returns application name
func (exe *Executor) GetName() string {
	return "Delete Volume " + exe.volumeName
}

// Execute deletes application by deleting its associated flow
func (exe *Executor) Execute() error {
	util.SpinStart("Deleting Volume")
	volume, err := exe.ns.GetVolume(exe.volumeName)
	if err != nil {
		return err
	}

	// Delete files
	ch := make(chan error, len(volume.Agents))
	for idx := range volume.Agents {
		go exe.execute(volume, idx, ch)
	}
	for idx := 0; idx < len(volume.Agents); idx++ {
		if err := <-ch; err != nil {
			return err
		}
	}

	// Delete from config
	if err := exe.ns.DeleteVolume(exe.volumeName); err != nil {
		return err
	}
	return config.Flush()
}

func (exe *Executor) execute(volume rsc.Volume, agentIdx int, ch chan error) {
	agentName := volume.Agents[agentIdx]
	baseAgent, err := exe.ns.GetAgent(agentName)
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
	if _, err := ssh.Run("rm -rf " + util.AddTrailingSlash(volume.Destination) + "*"); err != nil {
		ch <- err
	}

	ch <- nil
}
