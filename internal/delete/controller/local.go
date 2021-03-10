/*
 *  *******************************************************************************
 *  * Copyright (c) 2020 Edgeworx, Inc.
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

	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

type LocalExecutor struct {
	controlPlane          *rsc.LocalControlPlane
	namespace             string
	name                  string
	localControllerConfig *install.LocalContainerConfig
}

func NewLocalExecutor(controlPlane *rsc.LocalControlPlane, namespace, name string) *LocalExecutor {
	exe := &LocalExecutor{
		controlPlane:          controlPlane,
		namespace:             namespace,
		name:                  name,
		localControllerConfig: install.NewLocalControllerConfig("", install.Credentials{}),
	}
	return exe
}

func (exe *LocalExecutor) GetName() string {
	return exe.name
}

func (exe *LocalExecutor) Execute() error {
	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
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
	if err := ns.DeleteController(exe.name); err != nil {
		return err
	}
	ns.SetControlPlane(exe.controlPlane)
	return config.Flush()
}
