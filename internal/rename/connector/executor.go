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

package connector

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func Execute(namespace, name, newName string) error {
	// Check that Connector exists in current namespace
	_, err := config.GetConnector(namespace, name)
	if err != nil {
		return err
	}

	util.SpinStart(fmt.Sprintf("Renaming Connector %s", name))

	// Do a shallow rename of controller
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return err
	}
	cnct, err := config.GetConnector(namespace, name)
	if err != nil {
		return err
	}
	endpoint, _ := ns.ControlPlane.GetControllerEndpoint()
	clt, err := client.NewAndLogin(endpoint, ns.ControlPlane.IofogUser.Email, ns.ControlPlane.IofogUser.Password)
	if err != nil {
		return err
	}
	var host = cnct.Host
	if host == "" {
		host = "0.0.0.0"
	}
	err = clt.AddConnector(client.ConnectorInfo{
		Name:   newName,
		Domain: host,
		IP:     host,
	})
	if err != nil {
		return err
	}
	err = clt.DeleteConnector(name)
	if err != nil {
		return err
	}
	config.Flush()
	return nil
}
