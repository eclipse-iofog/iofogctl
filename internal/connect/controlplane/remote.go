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

package connectcontrolplane

import (
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type remoteExecutor struct {
	ctrlPlane rsc.RemoteControlPlane
	namespace string
}

func newRemoteExecutor(ctrlPlane rsc.RemoteControlPlane, namespace string) *remoteExecutor {
	r := &remoteExecutor{
		ctrlPlane: ctrlPlane,
		namespace: namespace,
	}
	return r
}

func (exe *remoteExecutor) GetName() string {
	return "Control Plane"
}

func (exe *remoteExecutor) Execute() (err error) {
	// Establish connection
	controllers := exe.ctrlPlane.GetControllers()
	if len(controllers) == 0 {
		return util.NewError("Control Plane in Namespace " + exe.namespace + " has no Controllers. Try deploying a Control Plane to this Namespace.")
	}
	endpoint, err := exe.ctrlPlane.GetEndpoint()
	if err != nil {
		return err
	}
	err = connect(exe.ctrlPlane, endpoint, exe.namespace)
	if err != nil {
		return err
	}

	err = config.UpdateControlPlane(exe.namespace, exe.ctrlPlane)
	if err != nil {
		return err
	}

	return config.Flush()
}
