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

package connectcontrolplane

import (
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog/install"
)

type kubernetesExecutor struct {
	ctrlPlane config.ControlPlane
	namespace string
}

func newKubernetesExecutor(ctrlPlane config.ControlPlane, namespace string) *kubernetesExecutor {
	return &kubernetesExecutor{
		ctrlPlane: ctrlPlane,
		namespace: namespace,
	}
}

func (exe *kubernetesExecutor) GetName() string {
	return "Control Plane"
}

func (exe *kubernetesExecutor) Execute() (err error) {
	// Instantiate Kubernetes cluster object
	k8s, err := install.NewKubernetes(exe.ctrlPlane.Kube.Config, exe.namespace)
	if err != nil {
		return err
	}

	// Check the resources exist in K8s namespace
	if err = k8s.ExistsInNamespace(exe.namespace); err != nil {
		return err
	}

	// Get Controller endpoint
	endpoint, err := k8s.GetControllerEndpoint()
	if err != nil {
		return err
	}

	// Establish connection
	err = connect(exe.ctrlPlane, endpoint, exe.namespace)
	if err != nil {
		return err
	}

	exe.ctrlPlane.Controllers[0].Endpoint = endpoint
	err = config.UpdateControlPlane(exe.namespace, exe.ctrlPlane)
	if err != nil {
		return err
	}

	return config.Flush()
}
