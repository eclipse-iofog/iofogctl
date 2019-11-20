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

package deletecontroller

import (
	"fmt"
	"github.com/eclipse-iofog/iofogctl/internal/config"

	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type localExecutor struct {
	namespace             string
	name                  string
	client                *install.LocalContainer
	localControllerConfig *install.LocalContainerConfig
}

func newLocalExecutor(namespace, name string, client *install.LocalContainer) *localExecutor {
	exe := &localExecutor{
		namespace:             namespace,
		name:                  name,
		client:                client,
		localControllerConfig: install.NewLocalControllerConfig("", install.Credentials{}),
	}
	return exe
}

func (exe *localExecutor) GetName() string {
	return exe.name
}

func (exe *localExecutor) Execute() error {
	// Get container config
	// Clean container
	if errClean := exe.client.CleanContainer(exe.localControllerConfig.ContainerName); errClean != nil {
		util.PrintNotify(fmt.Sprintf("Could not clean Controller container: %v", errClean))
	}

	// Update config
	if err := config.DeleteController(exe.namespace, exe.name); err != nil {
		return err
	}

	return nil
}
