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
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type remoteExecutor struct {
	namespace          string
	cnct               *config.Connector
	controllerEndpoint string
	iofogUser          config.IofogUser
}

func newRemoteExecutor(namespace string, cnct *config.Connector) *remoteExecutor {
	d := &remoteExecutor{}
	d.namespace = namespace
	d.cnct = cnct
	return d
}

func (exe *remoteExecutor) GetName() string {
	return exe.cnct.Name
}

func (exe *remoteExecutor) Execute() (err error) {
	// Get Control Plane
	controlPlane, err := config.GetControlPlane(exe.namespace)
	if err != nil || len(controlPlane.Controllers) == 0 {
		util.PrintError("You must deploy a Controller to a namespace before deploying any Connector")
		return
	}
	exe.controllerEndpoint = controlPlane.Controllers[0].Endpoint
	exe.iofogUser = controlPlane.IofogUser

	// Instantiate installer
	connectorOptions := &install.ConnectorOptions{
		Name:               exe.cnct.Name,
		User:               exe.cnct.User,
		Host:               exe.cnct.Host,
		Port:               exe.cnct.Port,
		PrivKeyFilename:    exe.cnct.KeyFile,
		Version:            exe.cnct.Version,
		PackageCloudToken:  exe.cnct.PackageCloudToken,
		ControllerEndpoint: exe.controllerEndpoint,
		IofogUser:          install.IofogUser(exe.iofogUser),
	}
	installer := install.NewConnector(connectorOptions)

	// Install Connector
	if err = installer.Install(); err != nil {
		return
	}

	// Update connector (its a pointer, this is returned to caller)
	exe.cnct.Endpoint = exe.cnct.Host + ":" + iofog.ConnectorPortString
	exe.cnct.Created = util.NowUTC()

	return nil
}
