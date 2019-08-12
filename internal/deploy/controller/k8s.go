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
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/client"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
)

type kubernetesExecutor struct {
	namespace string
	ctrl      config.Controller
}

func newKubernetesExecutor(namespace string, ctrl config.Controller) *kubernetesExecutor {
	k := &kubernetesExecutor{}
	k.namespace = namespace
	k.ctrl = ctrl
	return k
}

func (exe *kubernetesExecutor) GetName() string {
	return exe.ctrl.Name
}

func (exe *kubernetesExecutor) Execute() (err error) {
	// Update configuration before we try to deploy in case of failure
	exe.ctrl, err = prepareUserAndSaveConfig(exe.namespace, exe.ctrl)
	if err != nil {
		return
	}

	// Get Kubernetes installer
	installer, err := install.NewKubernetes(exe.ctrl.KubeConfig, exe.namespace)
	if err != nil {
		return
	}

	// Configure deploy
	if err = installer.SetImages(exe.ctrl.Images); err != nil {
		return err
	}
	installer.SetControllerIP(exe.ctrl.KubeControllerIP)

	// Create controller on cluster
	endpoint, err := installer.CreateController(client.User{
		Name:     exe.ctrl.IofogUser.Name,
		Surname:  exe.ctrl.IofogUser.Surname,
		Email:    exe.ctrl.IofogUser.Email,
		Password: exe.ctrl.IofogUser.Password,
	})
	if err != nil {
		return
	}

	// Update configuration
	exe.ctrl.Endpoint = endpoint
	if err = config.UpdateController(exe.namespace, exe.ctrl); err != nil {
		return
	}

	return config.Flush()
}
