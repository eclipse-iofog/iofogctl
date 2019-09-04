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
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
)

type kubernetesExecutor struct {
	namespace    string
	ctrl         *config.Controller
	controlPlane config.ControlPlane
}

func newKubernetesExecutor(namespace string, ctrl *config.Controller, controlPlane config.ControlPlane) *kubernetesExecutor {
	k := &kubernetesExecutor{}
	k.namespace = namespace
	k.ctrl = ctrl
	k.controlPlane = controlPlane
	return k
}

func (exe *kubernetesExecutor) GetName() string {
	return exe.ctrl.Name
}

func (exe *kubernetesExecutor) Execute() (err error) {
	// Get Kubernetes installer
	installer, err := install.NewKubernetes(exe.ctrl.KubeConfig, exe.namespace)
	if err != nil {
		return
	}

	// Configure deploy
	if err = installer.SetImages(exe.controlPlane.Images); err != nil {
		return err
	}
	if exe.ctrl.KubeControllerIP != "" {
		installer.SetControllerIP(exe.ctrl.KubeControllerIP)
	}

	replicas := 1
	if exe.ctrl.Replicas != 0 {
		replicas = exe.ctrl.Replicas
	}
	// Create controller on cluster
	if err = installer.CreateController(install.IofogUser(exe.controlPlane.IofogUser), replicas, install.Database(exe.controlPlane.Database)); err != nil {
		return
	}

	// Update controller (its a pointer, this is returned to caller)
	if exe.ctrl.Endpoint, err = installer.GetControllerEndpoint(); err != nil {
		return
	}

	return
}
