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
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type localExecutor struct {
	volume rsc.Volume
	ns     *rsc.Namespace
	agents []*rsc.LocalAgent
}

func (exe localExecutor) GetName() string {
	return "deploying Volume " + exe.volume.Name
}

func (exe localExecutor) Execute() error {
	util.SpinStart("Pushing volumes to Agents")
	// Transfer files
	nbAgents := len(exe.agents)
	ch := make(chan error, nbAgents)
	for idx := range exe.agents {
		go exe.execute(idx, ch)
	}
	for idx := 0; idx < nbAgents; idx++ {
		if err := <-ch; err != nil {
			return err
		}
	}
	// Update config
	exe.ns.UpdateVolume(exe.volume)
	return config.Flush()
}

func (exe localExecutor) execute(agentIdx int, ch chan error) {
	// Docker cp
	client, err := install.NewLocalContainerClient()
	if err != nil {
		ch <- err
	}

	if err = client.CopyToContainer(install.GetLocalContainerName("agent", false), exe.volume.Source, exe.volume.Destination); err != nil {
		ch <- err
	}

	ch <- nil
}
