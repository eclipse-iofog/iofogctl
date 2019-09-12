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
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
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
		User:            ctrl.User,
		Host:            ctrl.Host,
		Port:            ctrl.Port,
		PrivKeyFilename: ctrl.KeyFile,
	}
	installer := install.NewController(controllerOptions)

	// Stop Controller
	if err = installer.Stop(); err != nil {
		return err
	}

	// Update config
	if err = config.DeleteController(exe.namespace, exe.name); err != nil {
		return err
	}

	return nil
}
