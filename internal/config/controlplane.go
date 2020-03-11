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

package config

import (
	"strconv"

	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

// GetControllerEndpoint returns ioFog controller endpoint
func (ctrlPlane ControlPlane) GetControllerEndpoint() (string, error) {
	// Loadbalancer ?
	if ctrlPlane.LoadBalancer.Host != "" {
		if ctrlPlane.LoadBalancer.Port != 0 {
			return ctrlPlane.LoadBalancer.Host + ":" + strconv.Itoa(ctrlPlane.LoadBalancer.Port), nil
		}
		return ctrlPlane.LoadBalancer.Host + ":" + iofog.ControllerPortString, nil
	}

	// First controller
	if len(ctrlPlane.Controllers) < 1 {
		return "", util.NewError("This control plane does not have controller")
	}
	return ctrlPlane.Controllers[0].Endpoint, nil
}
