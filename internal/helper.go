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

package internal

import (
	"github.com/eclipse-iofog/iofog-go-sdk/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
)

// NewControllerClient returns an iofog-go-sdk/client configured for the current namespace
func NewControllerClient(namespace string) (clt *client.Client, err error) {
	// Get Control Plane
	controlPlane, err := config.GetControlPlane(namespace)
	if err != nil {
		return nil, err
	}
	endpoint, err := controlPlane.GetControllerEndpoint()
	if err != nil {
		return nil, err
	}
	return client.NewAndLogin(endpoint, controlPlane.IofogUser.Email, controlPlane.IofogUser.Password)
}
