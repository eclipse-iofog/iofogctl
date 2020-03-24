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
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog/install"
)

type kubernetesExecutor struct {
	controlPlane *rsc.KubernetesControlPlane
	namespace    string
	name         string
}

func newKubernetesExecutor(controlPlane *rsc.KubernetesControlPlane, namespace, name string) *kubernetesExecutor {
	return &kubernetesExecutor{
		controlPlane: controlPlane,
		namespace:    namespace,
		name:         name,
	}
}

func (exe *kubernetesExecutor) GetName() string {
	return exe.name
}

func (exe *kubernetesExecutor) Execute() error {
	// Instantiate Kubernetes object
	k8s, err := install.NewKubernetes(exe.controlPlane.KubeConfig, exe.namespace)

	// Delete Controller on cluster
	err = k8s.DeleteController()
	if err != nil {
		return err
	}

	// Update config
	if err = exe.controlPlane.DeleteController(exe.name); err != nil {
		return err
	}
	config.UpdateControlPlane(exe.namespace, exe.controlPlane)

	return nil
}
