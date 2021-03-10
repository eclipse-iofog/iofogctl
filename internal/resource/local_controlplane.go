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

import (
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

type LocalControlPlane struct {
	IofogUser  IofogUser        `yaml:"iofogUser"`
	Controller *LocalController `yaml:"controller,omitempty"`
}

func (cp *LocalControlPlane) GetUser() IofogUser {
	return cp.IofogUser
}

func (cp *LocalControlPlane) GetControllers() []Controller {
	if cp.Controller == nil {
		return []Controller{}
	}
	return []Controller{cp.Controller.Clone()}
}

func (cp *LocalControlPlane) GetController(name string) (Controller, error) {
	if cp.Controller == nil {
		return nil, util.NewError("Local Control Plane does not have a Controller")
	}
	return cp.Controller, nil
}

func (cp *LocalControlPlane) GetEndpoint() (string, error) {
	if cp.Controller == nil {
		return "", util.NewError("Local Control Plane does not have a Controller, cannot get endpoint.")
	}
	return cp.Controller.GetEndpoint(), nil
}

func (cp *LocalControlPlane) UpdateController(baseController Controller) error {
	controller, ok := baseController.(*LocalController)
	if !ok {
		return util.NewError("Must add Local Controller to Local Control Plane")
	}
	cp.Controller = controller
	return nil
}

func (cp *LocalControlPlane) AddController(baseController Controller) error {
	controller, ok := baseController.(*LocalController)
	if !ok {
		return util.NewError("Must add Local Controller to Local Control Plane")
	}
	cp.Controller = controller
	return nil
}

func (cp *LocalControlPlane) DeleteController(string) error {
	cp.Controller = nil
	return nil
}

func (cp *LocalControlPlane) Sanitize() error {
	if cp.Controller != nil && !util.IsLocalHost(cp.Controller.Endpoint) {
		cp.Controller.Endpoint = "localhost"
	}
	return nil
}

func (cp *LocalControlPlane) Clone() ControlPlane {
	return &LocalControlPlane{
		IofogUser:  cp.IofogUser,
		Controller: cp.Controller.Clone().(*LocalController),
	}
}
