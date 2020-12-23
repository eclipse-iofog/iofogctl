/*
 *  *******************************************************************************
 *  * Copyright (c) 2020 Edgeworx, Inc.
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package resource

type KubernetesController struct {
	PodName  string `yaml:"podName"`
	Endpoint string `yaml:"endpoint"`
	Created  string `yaml:"created,omitempty"`
	Status   string `yaml:"status,omitempty"`
}

func (ctrl *KubernetesController) GetName() string {
	return ctrl.PodName
}

func (ctrl *KubernetesController) GetEndpoint() string {
	return ctrl.Endpoint
}

func (ctrl *KubernetesController) GetCreatedTime() string {
	return ctrl.Created
}

func (ctrl *KubernetesController) SetName(name string) {
	ctrl.PodName = name
}

func (ctrl *KubernetesController) Sanitize() error {
	return nil
}

func (ctrl *KubernetesController) Clone() Controller {
	return &KubernetesController{
		PodName:  ctrl.PodName,
		Endpoint: ctrl.Endpoint,
		Created:  ctrl.Created,
	}
}
