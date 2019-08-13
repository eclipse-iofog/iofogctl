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

package deployconnector

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func NewExecutor(namespace string, cnct *config.Connector, controlPlane config.ControlPlane) (execute.Executor, error) {
	// Must contain Controller
	if len(controlPlane.Controllers) == 0 {
		return nil, util.NewError("There are no Controllers in this namespace. You must first deploy one or more Controllers.")
	}

	// Must contain an ioFog User
	if controlPlane.IofogUser.Email == "" || controlPlane.IofogUser.Password == "" {
		return nil, util.NewError("The Control Plane in this namespace does not have a valid ioFog user")
	}

	// Local executor
	if util.IsLocalHost(cnct.Host) {
		existingConnectors, err := config.GetConnectors(namespace)
		if err != nil {
			return nil, err
		}
		// Check the namespace does not contain a Connector yet
		nbConnectors := len(existingConnectors)
		if nbConnectors > 0 {
			return nil, util.NewInputError("This namespace already contains a local Connector. Please remove it before deploying a new one.")
		}
		return nil, util.NewError("Local Connector deploy functionality is not implemented yet")
	}

	// Default executor
	if cnct.Host == "" || cnct.KeyFile == "" || cnct.User == "" {
		return nil, util.NewInputError("Must specify user, host, and key file flags for remote deployment")
	}
	// TODO: Replace Controllers[0].Endpoint with different variable e.g. loadbalancer
	return newRemoteExecutor(namespace, cnct, controlPlane.Controllers[0].Endpoint, controlPlane.IofogUser), nil
}
