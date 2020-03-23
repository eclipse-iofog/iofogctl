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
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog/install"
)

type kubernetesExecutor struct {
	namespace    string
	controlPlane *rsc.KubernetesControlPlane
}

func newKubernetesExecutor(namespace string, controlPlane *rsc.ControlPlane) *kubernetesExecutor {
	return &kubernetesExecutor{
		namespace:    namespace,
		controlPlane: controlPlane,
	}
}

func (exe *kubernetesExecutor) GetName() string {
	return "Deploy Kubernetes Control Plane"
}

func (exe *kubernetesExecutor) Execute() (err error) {
	// Get Kubernetes deployer
	installer, err := install.NewKubernetes(exe.controlPlane.Kube.Config, exe.namespace)
	if err != nil {
		return
	}

	// Configure deploy
	installer.SetKubeletImage(exe.controlPlane.Kube.Images.Kubelet)
	installer.SetOperatorImage(exe.controlPlane.Kube.Images.Operator)
	installer.SetPortManagerImage(exe.controlPlane.Kube.Images.PortManager)
	installer.SetRouterImage(exe.controlPlane.Kube.Images.Router)
	installer.SetProxyImage(exe.controlPlane.Kube.Images.Proxy)
	installer.SetControllerImage(exe.controlPlane.Kube.Images.Controller)
	installer.SetControllerService(exe.controlPlane.Kube.Services.Controller.Type, exe.controlPlane.Kube.Services.Controller.IP)
	installer.SetRouterService(exe.controlPlane.Kube.Services.Router.Type, exe.controlPlane.Kube.Services.Router.IP)
	installer.SetRouterService(exe.controlPlane.Kube.Services.Proxy.Type, exe.controlPlane.Kube.Services.Proxy.IP)

	replicas := int32(1)
	if exe.controlPlane.Kube.Replicas.Controller != 0 {
		replicas = exe.controlPlane.Kube.Replicas.Controller
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
