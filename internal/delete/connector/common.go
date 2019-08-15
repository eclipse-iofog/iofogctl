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

package deleteconnector

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/client"
)

func deleteConnectorFromController(namespace, connectorIP string) error {
	// Get the Control Plane to access Controller API
	controlPlane, err := config.GetControlPlane(namespace)
	if err != nil {
		return err
	}
	if len(controlPlane.Controllers) == 0 {
		// No Controllers, finish
		return nil
	}
	// Login and delete the Connector
	// TODO: replace endpoint with controlplane var
	ctrlClient := client.New(controlPlane.Controllers[0].Endpoint)
	if err = ctrlClient.Login(client.LoginRequest{
		Email:    controlPlane.IofogUser.Email,
		Password: controlPlane.IofogUser.Password,
	}); err != nil {
		return err
	}
	if err = ctrlClient.DeleteConnector(connectorIP); err != nil {
		return err
	}

	return nil
}
