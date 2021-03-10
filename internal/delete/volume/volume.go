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
	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	"github.com/eclipse-iofog/iofogctl/v3/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
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
		go exe.execute(&volume, idx, ch)
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

// TODO: Parallelize this
func (exe *Executor) execute(volume *rsc.Volume, agentIdx int, ch chan error) {
	agentName := volume.Agents[agentIdx]
	baseAgent, err := exe.ns.GetAgent(agentName)
	if err != nil {
		ch <- err
	}
	agent, ok := baseAgent.(*rsc.RemoteAgent)
	if ok {
		if err = deleteRemote(agent, volume); err != nil {
			ch <- err
		}
	} else {
		agent, ok := baseAgent.(*rsc.LocalAgent)
		if ok {
			if err = deleteLocal(agent, volume); err != nil {
				ch <- err
			}
		} else {
			ch <- util.NewError("Could not convert Agent")
		}
	}

	ch <- nil
}
