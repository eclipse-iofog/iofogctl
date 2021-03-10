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
	"fmt"

	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

type localExecutor struct {
	volume rsc.Volume
	ns     *rsc.Namespace
	agents []*rsc.LocalAgent
}

func (exe *localExecutor) GetName() string {
	return "deploying Volume " + exe.volume.Name
}

func (exe *localExecutor) Execute() error {
	if len(exe.agents) == 0 {
		return nil
	}
	util.SpinStart("Pushing volumes to Agents")
	util.PrintNotify("Local Agent uses the host filesystem when mounting/binding volumes to the Microservices. Therefore deploying a Volume to a Local Agent is unecessary.")
	if exe.volume.Source != exe.volume.Destination {
		msg := `Source '%s' is different from destination '%s'
This may result cause issues, as the Microservices running on the Local Agent will use the host filesystem to bind/mount volumes.`
		util.PrintNotify(fmt.Sprintf(msg, exe.volume.Source, exe.volume.Destination))
	}
	// Update config
	exe.ns.UpdateVolume(&exe.volume)
	return config.Flush()
}
