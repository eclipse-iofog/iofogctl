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
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type remoteExecutor struct {
	namespace string
	name      string
}

func newRemoteExecutor(namespace, name string) *remoteExecutor {
	exe := &remoteExecutor{}
	exe.namespace = namespace
	exe.name = name
	return exe
}

func (exe *remoteExecutor) GetName() string {
	return exe.name
}

func (exe *remoteExecutor) Execute() error {
	// Get controller from config
	cnct, err := config.GetConnector(exe.namespace, exe.name)
	if err != nil {
		return err
	}

	// Instantiate installer
	connectorOptions := &install.ConnectorOptions{
		User:            cnct.SSH.User,
		Host:            cnct.Host,
		Port:            cnct.SSH.Port,
		PrivKeyFilename: cnct.SSH.KeyFile,
	}
	installer := install.NewConnector(connectorOptions)

	// Stop Connector
	if err = installer.Stop(); err != nil {
		util.PrintNotify("Could not stop daemon for Connector " + exe.name)
	}

	// Clear Connector from Controller
	if err = deleteConnectorFromController(exe.namespace, cnct.Host); err != nil {
		return err
	}

	// Update config
	if err := config.DeleteConnector(exe.namespace, exe.name); err != nil {
		return err
	}

	return nil
}
