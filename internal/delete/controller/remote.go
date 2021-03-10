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
	"github.com/eclipse-iofog/iofogctl/v3/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

type RemoteExecutor struct {
	controlPlane *rsc.RemoteControlPlane
	namespace    string
	name         string
}

func NewRemoteExecutor(controlPlane *rsc.RemoteControlPlane, namespace, name string) *RemoteExecutor {
	return &RemoteExecutor{
		controlPlane: controlPlane,
		namespace:    namespace,
		name:         name,
	}
}

func (exe *RemoteExecutor) GetName() string {
	return exe.name
}

func (exe *RemoteExecutor) Execute() error {
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

	// Try to remove default router
	sshAgent, err := install.NewRemoteAgent(ctrl.SSH.User,
		ctrl.Host,
		ctrl.SSH.Port,
		ctrl.SSH.KeyFile,
		iofog.VanillaRouterAgentName,
		"")
	if err != nil {
		return err
	}
	if err = sshAgent.Uninstall(); err != nil {
		util.PrintNotify(fmt.Sprintf("Failed to stop daemon on Agent %s. %s", iofog.VanillaRouterAgentName, err.Error()))
	}

	// Instantiate Controller uninstaller
	controllerOptions := &install.ControllerOptions{
		User:            ctrl.SSH.User,
		Host:            ctrl.Host,
		Port:            ctrl.SSH.Port,
		PrivKeyFilename: ctrl.SSH.KeyFile,
	}
	installer, err := install.NewController(controllerOptions)
	if err != nil {
		return err
	}

	// Uninstall Controller
	if err := installer.Uninstall(); err != nil {
		return err
	}

	// Update config
	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	if err := ns.DeleteController(exe.name); err != nil {
		return err
	}
	return config.Flush()
}
