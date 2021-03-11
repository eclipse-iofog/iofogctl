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

package describe

import (
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

func (exe *applicationExecutor) initLegacy() (err error) {
	exe.flow, err = exe.client.GetFlowByName(exe.name)
	if err != nil {
		return
	}
	msvcListResponse, err := exe.client.GetMicroservicesPerFlow(exe.flow.ID)
	if err != nil {
		return
	}

	// Filter system microservices
	for idx := range msvcListResponse.Microservices {
		msvc := &msvcListResponse.Microservices[idx]
		if util.IsSystemMsvc(msvc) {
			continue
		}
		exe.msvcs = append(exe.msvcs, msvc)
	}
	exe.msvcPerID = make(map[string]*client.MicroserviceInfo)
	for i := 0; i < len(exe.msvcs); i++ {
		exe.msvcPerID[exe.msvcs[i].UUID] = exe.msvcs[i]
	}
	return
}
