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
	"fmt"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func (exe executor) remoteRemove() error {
	// Get controller from config
	cnct, err := config.GetConnector(exe.namespace, exe.name)
	if err != nil {
		return err
	}

	if cnct.Host == "" || cnct.SSH.User == "" || cnct.SSH.KeyFile == "" || cnct.SSH.Port == 0 {
		util.PrintNotify("Could not stop daemon for Connector " + cnct.Name + ". SSH details missing from local cofiguration. Use configure command to add SSH details.")
	} else {
		// Instantiate installer
		installer := install.NewConnector(&install.ConnectorOptions{
			User:            cnct.SSH.User,
			Host:            cnct.Host,
			Port:            cnct.SSH.Port,
			PrivKeyFilename: cnct.SSH.KeyFile,
		})

		// Stop Connector
		if err = installer.Uninstall(); err != nil {
			util.PrintNotify(fmt.Sprintf("Failed to stop daemon on Connector %s. Error: %s", cnct.Name, err.Error()))
		}
	}

	// Update config
	if err := config.DeleteConnector(exe.namespace, exe.name); err != nil {
		return err
	}

	return nil
}
