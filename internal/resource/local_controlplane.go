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
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type LocalControlPlane struct {
	IofogUser  IofogUser        `yaml:"iofogUser"`
	Controller *LocalController `yaml:"controller,omitempty"`
}

func (cp LocalControlPlane) GetUser() IofogUser {
	return cp.IofogUser
}

func (cp LocalControlPlane) GetControllers() []Controller {
	return []Controller{cp.Controller}
}

func (cp LocalControlPlane) GetController(name string) (Controller, error) {
	return cp.Controller, nil
}

func (cp LocalControlPlane) GetEndpoint() (string, error) {
	return cp.Controller.GetEndpoint(), nil
}

func (cp *LocalControlPlane) UpdateController(baseController Controller) error {
	controller, ok := baseController.(*LocalController)
	if !ok {
		return util.NewError("Could not convert Controller to Local Controller")
	}
	cp.Controller = controller
	return nil
}

func (cp *LocalControlPlane) AddController(baseController Controller) error {
	controller, ok := baseController.(*LocalController)
	if !ok {
		return util.NewError("Could not convert Controller to Local Controller")
	}
	cp.Controller = controller
	return nil
}

func (cp *LocalControlPlane) DeleteController(string) error {
	cp.Controller = &LocalController{}
	return nil
}

func (cp *LocalControlPlane) Sanitize() error {
	// Nothing to sanitize
	return nil
}
