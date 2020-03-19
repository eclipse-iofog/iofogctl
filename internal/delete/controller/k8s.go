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

package deletecontroller

import (
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog/install"
)

type kubernetesExecutor struct {
	namespace string
	name      string
}

func newKubernetesExecutor(namespace, name string) *kubernetesExecutor {
	return &kubernetesExecutor{
		namespace: namespace,
		name:      name,
	}
}

func (exe *kubernetesExecutor) GetName() string {
	return exe.name
}

func (exe *kubernetesExecutor) Execute() error {
	// Get controller from config
	controlPlane, err := config.GetControlPlane(exe.namespace)
	if err != nil {
		return err
	}

	// Instantiate Kubernetes object
	k8s, err := install.NewKubernetes(controlPlane.Kube.Config, exe.namespace)

	// Delete Controller on cluster
	err = k8s.DeleteController()
	if err != nil {
		return err
	}

	// Update config
	if err = config.DeleteController(exe.namespace, exe.name); err != nil {
		return err
	}

	return nil
}
