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
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"

	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type localExecutor struct {
	controlPlane          *rsc.LocalControlPlane
	namespace             string
	name                  string
	client                *install.LocalContainer
	localControllerConfig *install.LocalContainerConfig
}

func newLocalExecutor(controlPlane *rsc.LocalControlPlane, namespace, name string) *localExecutor {
	exe := &localExecutor{
		controlPlane:          controlPlane,
		namespace:             namespace,
		name:                  name,
		localControllerConfig: install.NewLocalControllerConfig("", install.Credentials{}),
	}
	return exe
}

func (exe *localExecutor) GetName() string {
	return exe.name
}

func (exe *localExecutor) Execute() error {
	client, err := install.NewLocalContainerClient()
	if err != nil {
		return err
	}
	// Get container config
	// Clean container
	if errClean := client.CleanContainer(exe.localControllerConfig.ContainerName); errClean != nil {
		util.PrintNotify(fmt.Sprintf("Could not clean Controller container: %v", errClean))
	}

	// Update config
	if err := exe.controlPlane.DeleteController(exe.name); err != nil {
		return err
	}
	config.UpdateControlPlane(exe.namespace, exe.controlPlane)

	return nil
}
