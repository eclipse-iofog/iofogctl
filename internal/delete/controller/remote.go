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

	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type remoteExecutor struct {
	controlPlane *rsc.RemoteControlPlane
	namespace    string
	name         string
}

func newRemoteExecutor(controlPlane *rsc.RemoteControlPlane, namespace, name string) *remoteExecutor {
	return &remoteExecutor{
		controlPlane: controlPlane,
		namespace:    namespace,
		name:         name,
	}
}

func (exe *remoteExecutor) GetName() string {
	return exe.name
}

func (exe *remoteExecutor) Execute() error {
	// Get controller from config
	baseCtrl, err := exe.controlPlane.GetController(exe.name)
	if err != nil {
		return err
	}

	// Assert dynamic type
	ctrl, ok := baseCtrl.(*rsc.RemoteController)
	if !ok {
		return util.NewInternalError("Could not assert Controller type to Remote Controller")
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
		"")
	if err = sshAgent.Uninstall(); err != nil {
		util.PrintNotify(fmt.Sprintf("Failed to stop daemon on Agent %s. %s", iofog.VanillaRouterAgentName, err.Error()))
	}

	// Update config
	if err = exe.controlPlane.DeleteController(exe.namespace); err != nil {
		return err
	}

	return nil
}
