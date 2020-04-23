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
	exe.ns.UpdateVolume(exe.volume)
	return config.Flush()
}

func (exe localExecutor) execute(agentIdx int, ch chan error) {
	// agent := exe.agents[agentIdx]

	// Docker cp

	ch <- nil
}
