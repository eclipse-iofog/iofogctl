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
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
)

type kubernetesExecutor struct {
	ctrlPlane config.ControlPlane
	namespace string
}

func newKubernetesExecutor(ctrlPlane config.ControlPlane, namespace string) *kubernetesExecutor {
	k := &kubernetesExecutor{
		ctrlPlane: ctrlPlane,
		namespace: namespace,
	}
	return k
}

func (exe *kubernetesExecutor) GetName() string {
	return "Control Plane"
}

func (exe *kubernetesExecutor) Execute() (err error) {
	// Instantiate Kubernetes cluster object
	k8s, err := install.NewKubernetes(exe.ctrlPlane.Controllers[0].KubeConfig, exe.namespace)
	if err != nil {
		return err
	}

	// Check the resources exist in K8s namespace
	err = k8s.ExistsInNamespace(exe.opt.Namespace)
	if err != nil {
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

	err = config.UpdateControlPlane(exe.namespace, exe.ctrlPlane)
	if err != nil {
		return err
	}

	return config.Flush()
}
