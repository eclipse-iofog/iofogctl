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

package resource

import (
	"fmt"
)

type Controller interface {
	GetName() string
	GetEndpoint() string
	SetName(string)
}

type LocalController struct {
	Name      string    `yaml:"name"`
	Endpoint  string    `yaml:"endpoint"`
	Container Container `yaml:"container"`
	Created   string    `yaml:"created,omitempty"`
}

type KubernetesController struct {
	PodName string `yaml:"podName"`
	Created string `yaml:"created,omitempty"`
}

type RemoteController struct {
	Name        string  `yaml:"name"`
	Host        string  `yaml:"host"`
	Port        int32   `yaml:"port"`
	SSH         SSH     `yaml:"ssh,omitempty"`
	Endpoint    string  `yaml:"endpoint,omitempty"`
	Created     string  `yaml:"created,omitempty"`
	Package     Package `yaml:"package,omitempty"`
	SystemAgent Package `yaml:"systemAgent,omitempty"`
}

func (ctrl LocalController) GetName() string {
	return ctrl.Name
}

func (ctrl LocalController) GetEndpoint() string {
	return ctrl.Endpoint
}

func (ctrl *LocalController) SetName(name string) {
	ctrl.Name = name
}

func (ctrl KubernetesController) GetName() string {
	return ctrl.PodName
}

func (ctrl KubernetesController) GetEndpoint() string {
	return ""
}

func (ctrl *KubernetesController) SetName(name string) {
	ctrl.PodName = name
}

func (ctrl RemoteController) GetName() string {
	return ctrl.Name
}

func (ctrl RemoteController) GetEndpoint() string {
	return fmt.Sprintf("%s:%d", ctrl.Host, ctrl.Port)
}

func (ctrl *RemoteController) SetName(name string) {
	ctrl.Name = name
}
