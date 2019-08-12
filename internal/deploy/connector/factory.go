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

func NewExecutor(namespace string, cnct config.Connector) (execute.Executor, error) {
	// Get the namespace
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return nil, err
	}

	// Must contain Controller
	if len(ns.ControlPlane.Controllers) == 0 {
		return nil, util.NewError("There are no Controllers in this namespace. You must first deploy one or more Controllers.")
	}

	// Local executor
	if util.IsLocalHost(cnct.Host) {
		// Check the namespace does not contain a Connector yet
		nbConnectors := len(ns.ControlPlane.Connectors)
		if nbConnectors > 0 {
			return nil, util.NewInputError("This namespace already contains a local Connector. Please remove it before deploying a new one.")
		}
		return nil, util.NewError("Local Connector deploy functionality is not implemented yet")
	}

	// Default executor
	if cnct.Host == "" || cnct.KeyFile == "" || cnct.User == "" {
		return nil, util.NewInputError("Must specify user, host, and key file flags for remote deployment")
	}
	ctrl := ns.ControlPlane.Controllers[0]
	return newRemoteExecutor(namespace, cnct, ctrl.Endpoint, ctrl.IofogUser), nil
}
