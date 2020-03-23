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
	"fmt"

	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type kubernetesExecutor struct {
	controlPlane *rsc.KubernetesControlPlane
	namespace    string
}

func newKubernetesExecutor(controlPlane *rsc.KubernetesControlPlane, namespace string) *kubernetesExecutor {
	return &kubernetesExecutor{
		controlPlane: controlPlane,
		namespace:    namespace,
	}
}

func (exe *kubernetesExecutor) GetName() string {
	return "Control Plane"
}

func (exe *kubernetesExecutor) Execute() (err error) {
	// Instantiate Kubernetes cluster object
	k8s, err := install.NewKubernetes(exe.controlPlane.Kube.Config, exe.namespace)
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
	err = connect(exe.controlPlane, endpoint, exe.namespace)
	if err != nil {
		return err
	}

	// TODO: Get Kubernetes pods
	for idx := int32(0); idx < exe.controlPlane.Replicas.Controller; idx++ {
		k8sPod := rsc.KubernetesController{
			Endpoint: endpoint,
			PodName:  fmt.Sprintf("Kubernetes-%d", idx),
			Created:  util.NowUTC(),
		}
		if err := exe.controlPlane.AddController(&k8sPod); err != nil {
			return err
		}
	}
	err = config.UpdateControlPlane(exe.namespace, exe.controlPlane)
	if err != nil {
		return err
	}

	return config.Flush()
}
