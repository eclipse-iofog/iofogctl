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
	return &kubernetesExecutor{
		namespace:    namespace,
		ctrl:         ctrl,
		controlPlane: controlPlane,
	}
}

func (exe *kubernetesExecutor) GetName() string {
	return exe.ctrl.Name
}

func (exe *kubernetesExecutor) Execute() (err error) {
	// Get Kubernetes deployer
	installer, err := install.NewKubernetes(exe.ctrl.Kube.Config, exe.namespace)
	if err != nil {
		return
	}

	// Configure deploy
	installer.SetKubeletImage(exe.ctrl.Kube.Images.Kubelet)
	installer.SetOperatorImage(exe.ctrl.Kube.Images.Operator)
	installer.SetPortManagerImage(exe.ctrl.Kube.Images.PortManager)
	installer.SetControllerImage(exe.ctrl.Container.Image)
	installer.SetControllerIP(exe.ctrl.Kube.StaticIP)
	if err = installer.SetControllerServiceType(exe.ctrl.Kube.ServiceType); err != nil {
		return
	}

	replicas := 1
	if exe.ctrl.Kube.Replicas != 0 {
		replicas = exe.ctrl.Kube.Replicas
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
