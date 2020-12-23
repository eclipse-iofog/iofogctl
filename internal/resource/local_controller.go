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

type LocalController struct {
	Name      string    `yaml:"name"`
	Endpoint  string    `yaml:"endpoint"`
	Container Container `yaml:"container"`
	Created   string    `yaml:"created,omitempty"`
}

func (ctrl *LocalController) GetName() string {
	return ctrl.Name
}

func (ctrl *LocalController) GetEndpoint() string {
	return ctrl.Endpoint
}

func (ctrl *LocalController) GetCreatedTime() string {
	return ctrl.Created
}

func (ctrl *LocalController) SetName(name string) {
	ctrl.Name = name
}

func (ctrl *LocalController) Sanitize() error {
	if ctrl.Name == "" {
		ctrl.Name = "local"
	}
	return nil
}

func (ctrl *LocalController) Clone() Controller {
	return &LocalController{
		Name:      ctrl.Name,
		Endpoint:  ctrl.Endpoint,
		Container: ctrl.Container,
		Created:   ctrl.Created,
	}
}
