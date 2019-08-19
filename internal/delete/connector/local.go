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

	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type localExecutor struct {
	namespace            string
	name                 string
	client               *install.LocalContainer
	localConnectorConfig *install.LocalContainerConfig
}

func newLocalExecutor(namespace, name string, client *install.LocalContainer) *localExecutor {
	exe := &localExecutor{
		namespace:            namespace,
		name:                 name,
		client:               client,
		localConnectorConfig: install.NewLocalConnectorConfig("", install.Credentials{}),
	}
	return exe
}

func (exe *localExecutor) GetName() string {
	return exe.name
}

func (exe *localExecutor) Execute() error {
	// Clean container
	if errClean := exe.client.CleanContainer(exe.localConnectorConfig.ContainerName); errClean != nil {
		util.PrintNotify(fmt.Sprintf("Could not clean Connector container: %v", errClean))
	}

	// Clear Connector from Controller
	if err := deleteConnectorFromController(exe.namespace, exe.localConnectorConfig.Host); err != nil {
		return err
	}

	return nil
}
