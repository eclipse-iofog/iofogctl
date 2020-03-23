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
	"fmt"

	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type kubernetesExecutor struct {
	namespace    string
	controlPlane *rsc.KubernetesControlPlane
}

func newKubernetesExecutor(namespace string, controlPlane *rsc.KubernetesControlPlane) *kubernetesExecutor {
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
	installer, err := install.NewKubernetes(exe.controlPlane.KubeConfig, exe.namespace)
	if err != nil {
		return
	}

	// Configure deploy
	installer.SetKubeletImage(exe.controlPlane.Images.Kubelet)
	installer.SetOperatorImage(exe.controlPlane.Images.Operator)
	installer.SetPortManagerImage(exe.controlPlane.Images.PortManager)
	installer.SetRouterImage(exe.controlPlane.Images.Router)
	installer.SetProxyImage(exe.controlPlane.Images.Proxy)
	installer.SetControllerImage(exe.controlPlane.Images.Controller)
	installer.SetControllerService(exe.controlPlane.Services.Controller.Type, exe.controlPlane.Services.Controller.IP)
	installer.SetRouterService(exe.controlPlane.Services.Router.Type, exe.controlPlane.Services.Router.IP)
	installer.SetRouterService(exe.controlPlane.Services.Proxy.Type, exe.controlPlane.Services.Proxy.IP)

	replicas := int32(1)
	if exe.controlPlane.Replicas.Controller != 0 {
		replicas = exe.controlPlane.Replicas.Controller
	}
	// Create controller on cluster
	if err = installer.CreateController(install.IofogUser(exe.controlPlane.IofogUser), replicas, install.Database(exe.controlPlane.Database)); err != nil {
		return
	}

	for idx := int32(0); idx < exe.controlPlane.Replicas.Controller; idx++ {
		if err := exe.controlPlane.AddController(&rsc.KubernetesController{
			PodName: fmt.Sprintf("kubernetes-%d", idx), // TODO: use actual pod name
			Created: util.NowUTC(),
		}); err != nil {
			return err
		}
	}
	// Update controller (its a pointer, this is returned to caller)
	if exe.controlPlane.Endpoint, err = installer.GetControllerEndpoint(); err != nil {
		return
	}

	return
}
