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
)

type Controller interface {
	GetName() string
	GetEndpoint() string
	SetName(string)
	SetEndpoint(string, int32)
}

type LocalController struct {
	name string `yaml:"name"`
	host string `yaml:"host"`
	port int32 `yaml:"port"`
	container Container `yaml:"container"`
}

type KubernetesController struct {
	name string `yaml:"name"`
	host string `yaml:"host"`
	port int32 `yaml:"port"`
}

type RemoteController struct {
	name string `yaml:"name"`
	host string `yaml:"host"`
	port int32 `yaml:"port"`
    SSH         SSH     `yaml:"ssh,omitempty"`
    Endpoint    string  `yaml:"endpoint,omitempty"`
    Created     string  `yaml:"created,omitempty"`
    Package     Package `yaml:"package,omitempty"`
    SystemAgent Package `yaml:"systemAgent,omitempty"`
}

func (ctrl LocalController) GetName() string {
	return ctrl.name
}

func (ctrl LocalController) GetEndpoint() string {
	return fmt.Sprintf("%s:%d", ctrl.host, ctrl.port)
}

func (ctrl *LocalController) SetName(name string) {
	ctrl.name = name
}

func (ctrl *LocalController) SetEndpoint(host string, port int32) {
	ctrl.host = host
	ctrl.port = port
}

func (ctrl KubernetesController) GetName() string {
	return ctrl.name
}

func (ctrl KubernetesController) GetEndpoint() string {
	return fmt.Sprintf("%s:%d", ctrl.host, ctrl.port)
}

func (ctrl *KubernetesController) SetName(name string) {
	ctrl.name = name
}

func (ctrl *KubernetesController) SetEndpoint(host string, port int32) {
	ctrl.host = host
	ctrl.port = port
}

func (ctrl RemoteController) GetName() string {
	return ctrl.name
}

func (ctrl RemoteController) GetEndpoint() string {
	return fmt.Sprintf("%s:%d", ctrl.host, ctrl.port)
}

func (ctrl *RemoteController) SetName(name string) {
	ctrl.name = name
}

func (ctrl *RemoteController) SetEndpoint(host string, port int32) {
	ctrl.host = host
	ctrl.port = port
}
