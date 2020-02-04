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
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
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
	ctrl, err := config.GetController(exe.namespace, exe.name)
	if err != nil {
		return err
	}

	// Instantiate installer
	controllerOptions := &install.ControllerOptions{
		User:            ctrl.SSH.User,
		Host:            ctrl.Host,
		Port:            ctrl.SSH.Port,
		PrivKeyFilename: ctrl.SSH.KeyFile,
	}
	installer := install.NewController(controllerOptions)

	// Uninstall Controller
	if err = installer.Uninstall(); err != nil {
		return err
	}

	// Try to remove default router
	sshAgent := install.NewRemoteAgent(
		ctrl.SSH.User,
		ctrl.Host,
		ctrl.SSH.Port,
		ctrl.SSH.KeyFile,
		iofog.VanillaRouterAgentName,
		"",
		nil)
	if err = sshAgent.Uninstall(); err != nil {
		util.PrintNotify(fmt.Sprintf("Failed to stop daemon on Agent %s. %s", iofog.VanillaRouterAgentName, err.Error()))
	}

	// Update config
	if err = config.DeleteController(exe.namespace, exe.name); err != nil {
		return err
	}

	return nil
}
