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

package deploycontroller

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
)

type remoteExecutor struct {
	namespace    string
	ctrl         *config.Controller
	controlPlane config.ControlPlane
}

func newRemoteExecutor(namespace string, ctrl *config.Controller, controlPlane config.ControlPlane) *remoteExecutor {
	d := &remoteExecutor{}
	d.namespace = namespace
	d.ctrl = ctrl
	d.controlPlane = controlPlane
	return d
}

func (exe *remoteExecutor) GetName() string {
	return exe.ctrl.Name
}

func (exe *remoteExecutor) Execute() (err error) {
	// Instantiate installer
	controllerOptions := &install.ControllerOptions{
		User:            exe.ctrl.User,
		Host:            exe.ctrl.Host,
		Port:            exe.ctrl.Port,
		PrivKeyFilename: exe.ctrl.KeyFile,
		Version:         exe.ctrl.Version,
		Repo:            exe.ctrl.Repo,
		Token:           exe.ctrl.Token,
	}
	installer := install.NewController(controllerOptions)

	// Set database configuration
	if exe.controlPlane.Database.Host != "" {
		db := exe.controlPlane.Database
		installer.SetControllerExternalDatabase(db.Host, db.User, db.Password, db.Port)
	}

	// Install Controller
	if err = installer.Install(); err != nil {
		return
	}
	// Update controller (its a pointer, this is returned to caller)
	exe.ctrl.Endpoint = exe.ctrl.Host + ":" + iofog.ControllerPortString

	return
}
