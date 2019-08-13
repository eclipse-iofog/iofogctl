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
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type remoteExecutor struct {
	namespace    string
	ctrl         config.Controller
	controlPlane config.ControlPlane
}

func newRemoteExecutor(namespace string, ctrl config.Controller, controlPlane config.ControlPlane) *remoteExecutor {
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
	defer util.SpinStop()
	util.SpinStart("Deploying Controller " + exe.ctrl.Name)

	// Instantiate installer
	controllerOptions := &install.ControllerOptions{
		User:              exe.ctrl.User,
		Host:              exe.ctrl.Host,
		Port:              exe.ctrl.Port,
		PrivKeyFilename:   exe.ctrl.KeyFile,
		Version:           exe.ctrl.Version,
		PackageCloudToken: exe.ctrl.PackageCloudToken,
	}
	installer := install.NewController(controllerOptions)

	// Install Controller
	if err = installer.Install(); err != nil {
		return
	}

	// TODO: This creates a race condition, but I can't relocate it
	// Update configuration
	exe.ctrl.Endpoint = exe.ctrl.Host + ":" + iofog.ControllerPortString
	if err = config.UpdateController(exe.namespace, exe.ctrl); err != nil {
		return
	}

	return
}
