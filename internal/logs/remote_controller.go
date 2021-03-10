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

package logs

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

type remoteControllerExecutor struct {
	controlPlane *rsc.RemoteControlPlane
	namespace    string
	name         string
}

func newRemoteControllerExecutor(controlPlane *rsc.RemoteControlPlane, namespace, name string) *remoteControllerExecutor {
	return &remoteControllerExecutor{
		controlPlane: controlPlane,
		namespace:    namespace,
		name:         name,
	}
}

func (exe *remoteControllerExecutor) GetName() string {
	return exe.name
}

func (exe *remoteControllerExecutor) Execute() error {
	// Get controller config
	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	controlPlane, err := ns.GetControlPlane()
	if err != nil {
		return err
	}
	baseCtrl, err := controlPlane.GetController(exe.name)
	if err != nil {
		return err
	}

	ctrl, ok := baseCtrl.(*rsc.RemoteController)
	if !ok {
		return util.NewInternalError("Could not assert Controller type to Remote Controller")
	}

	// Remote
	if err := ctrl.ValidateSSH(); err != nil {
		return err
	}
	ssh, err := util.NewSecureShellClient(ctrl.SSH.User, ctrl.Host, ctrl.SSH.KeyFile)
	if err != nil {
		return err
	}
	ssh.SetPort(ctrl.SSH.Port)
	if err := ssh.Connect(); err != nil {
		return err
	}

	// Get logs
	out, err := ssh.Run("sudo cat /var/log/iofog-controller/*")
	if err != nil {
		return err
	}
	fmt.Print(out.String())

	return nil
}
