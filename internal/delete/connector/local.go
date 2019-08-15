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
	namespace             string
	name                  string
	client                *install.LocalContainer
	localControllerConfig *install.LocalControllerConfig
}

func newLocalExecutor(namespace, name string, client *install.LocalContainer) *localExecutor {
	exe := &localExecutor{
		namespace:             namespace,
		name:                  name,
		client:                client,
		localControllerConfig: install.NewLocalControllerConfig(make(map[string]string)),
	}
	return exe
}

func (exe *localExecutor) GetName() string {
	return exe.name
}

func (exe *localExecutor) Execute() error {
	// Get container config
	containerConfig, exists := exe.localControllerConfig.ContainerMap["connector"]
	if !exists {
		return util.NewInternalError("Could not retrieve Connector container config")
	}
	// Clean container
	if errClean := exe.client.CleanContainer(containerConfig.ContainerName); errClean != nil {
		fmt.Printf("Could not clean Controller container: %v", errClean)
	}

	// Clear Connector from Controller
	if err := deleteConnectorFromController(exe.namespace, containerConfig.Host); err != nil {
		return err
	}

	return nil
}
