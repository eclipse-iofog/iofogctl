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
	util.PrintNotify("Local Agent uses the host filesystem when mounting/binding volumes to the microservices. Therefore deploying a Volume to a Local Agent is unecessary.")
	if exe.volume.Source != exe.volume.Destination {
		util.PrintNotify(fmt.Sprintf("[WARNING]: Source [%s] is different from destination [&s]\nThis may result in a bug, as the microservices running on the Local Agent will use the host filesystem to bind/mount volumes.", exe.volume.Source, exe.volume.Destination))
	}
	// Update config
	exe.ns.UpdateVolume(exe.volume)
	return config.Flush()
}
