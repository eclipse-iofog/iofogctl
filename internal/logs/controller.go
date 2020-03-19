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

package logs

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type controllerExecutor struct {
	namespace string
	name      string
}

func newControllerExecutor(namespace, name string) *controllerExecutor {
	return &controllerExecutor{
		namespace: namespace,
		name:      name,
	}
}

func (ctrl *controllerExecutor) GetName() string {
	return ctrl.name
}

func (exe *controllerExecutor) Execute() error {
	// Get controller config
	controlPlane, err := config.GetControlPlane(exe.namespace)
	if err != nil {
		return err
	}
	ctrl, err := controlPlane.GetController(exe.name)
	if err != nil {
		return err
	}

	// Local
	if util.IsLocalHost(ctrl.Host) {
		lc, err := install.NewLocalContainerClient()
		if err != nil {
			return err
		}
		containerName := install.GetLocalContainerName("controller", false)
		stdout, stderr, err := lc.GetLogsByName(containerName)
		if err != nil {
			return err
		}

		printContainerLogs(stdout, stderr)

		return nil
	}

	// K8s
	if controlPlane.Kube.Config != "" {
		out, err := util.Exec("KUBECONFIG="+controlPlane.Kube.Config, "kubectl", "logs", "-l", "name=controller", "-n", exe.namespace)
		if err != nil {
			return err
		}
		fmt.Print(out.String())
		return nil
	}

	// Remote
	if ctrl.Host == "" || ctrl.SSH.User == "" || ctrl.SSH.KeyFile == "" || ctrl.SSH.Port == 0 {
		util.Check(util.NewNoConfigError("Controller"))
	}
	ssh := util.NewSecureShellClient(ctrl.SSH.User, ctrl.Host, ctrl.SSH.KeyFile)
	ssh.SetPort(ctrl.SSH.Port)
	if err = ssh.Connect(); err != nil {
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
