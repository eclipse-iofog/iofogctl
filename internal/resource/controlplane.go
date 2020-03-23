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

type ControlPlane interface {
	GetUser() IofogUser
	GetControllers() []Controller
	GetController(string) (Controller, error)
	GetEndpoint() (string, error)
	UpdateController(Controller) error
	AddController(Controller) error
	DeleteController(string) error
}

type LocalControlPlane struct {
	IofogUser  IofogUser  `yaml:"iofogUser,omitempty"`
	Controller Controller `yaml:"controller,omitempty"`
}

type KubernetesControlPlane struct {
	ControllerPods []Controller `yaml:"controllerPods,omitempty"`
	Database       Database     `yaml:"database,omitempty"`
	IofogUser      IofogUser    `yaml:"iofogUser,omitempty"`
	Endpoint       string       `yaml:"endpoint,omitempty"`
	KubeConfig     string       `yaml:"config,omitempty"`
	Services       Services     `yaml:"services,omitempty"`
	Replicas       Replicas     `yaml:"replicas,omitempty"`
	Images         KubeImages   `yaml:"images,omitempty"`
}

type RemoteControlPlane struct {
	Database    Database     `yaml:"database,omitempty"`
	IofogUser   IofogUser    `yaml:"iofogUser,omitempty"`
	Controllers []Controller `yaml:"controllers,omitempty"`
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

func (cp *LocalControlPlane) UpdateController(ctrl Controller) error {
	cp.Controller = ctrl
	return nil
}

func (cp *LocalControlPlane) AddController(ctrl Controller) error {
	cp.Controller = ctrl
	return nil
}

func (cp *LocalControlPlane) DeleteController(string) error {
	cp.Controller = &LocalController{}
	return nil
}

func (cp KubernetesControlPlane) GetUser() IofogUser {
	return cp.IofogUser
}

func (cp KubernetesControlPlane) GetControllers() []Controller {
	return cp.ControllerPods
}

func (cp KubernetesControlPlane) GetController(name string) (ret Controller, err error) {
	for _, ctrl := range cp.ControllerPods {
		if ctrl.GetName() == name {
			ret = ctrl
			return
		}
	}
	err = util.NewError("Could not find Controller " + name)
	return
}

func (cp KubernetesControlPlane) GetEndpoint() (string, error) {
	return cp.Endpoint, nil
}

func (cp *KubernetesControlPlane) UpdateController(ctrl Controller) error {
	for idx := range cp.ControllerPods {
		if cp.ControllerPods[idx].GetName() == ctrl.GetName() {
			cp.ControllerPods[idx] = ctrl
			return nil
		}
	}
	cp.ControllerPods = append(cp.ControllerPods, ctrl)
	return nil
}

func (cp *KubernetesControlPlane) AddController(ctrl Controller) error {
	if _, err := cp.GetController(ctrl.GetName()); err == nil {
		return util.NewError("Could not add Controller " + ctrl.GetName() + " because it already exists")
	}
	cp.ControllerPods = append(cp.ControllerPods, ctrl)
	return nil
}

func (cp *KubernetesControlPlane) DeleteController(name string) error {
	for idx := range cp.ControllerPods {
		if cp.ControllerPods[idx].GetName() == name {
			cp.ControllerPods = append(cp.ControllerPods[:idx-1], cp.ControllerPods[idx+1:]...)
			return nil
		}
	}
	return util.NewError("Could not find Controller " + name + " when performing deletion")
}

func (cp RemoteControlPlane) GetUser() IofogUser {
	return cp.IofogUser
}

func (cp RemoteControlPlane) GetControllers() []Controller {
	return cp.Controllers
}

func (cp RemoteControlPlane) GetController(name string) (ret Controller, err error) {
	for _, ctrl := range cp.Controllers {
		if ctrl.GetName() == name {
			ret = ctrl
			return
		}
	}
	err = util.NewError("Could not find Controller " + name)
	return
}

func (cp RemoteControlPlane) GetEndpoint() (string, error) {
	if len(cp.Controllers) == 0 {
		return "", util.NewError("Control Plane does not have any Controllers")
	}
	return cp.GetControllers()[0].GetEndpoint(), nil
}

func (cp *RemoteControlPlane) UpdateController(ctrl Controller) error {
	for idx := range cp.Controllers {
		if cp.Controllers[idx].GetName() == ctrl.GetName() {
			cp.Controllers[idx] = ctrl
			return nil
		}
	}
	cp.Controllers = append(cp.Controllers, ctrl)
	return nil
}

func (cp *RemoteControlPlane) AddController(ctrl Controller) error {
	if _, err := cp.GetController(ctrl.GetName()); err == nil {
		return util.NewError("Could not add Controller " + ctrl.GetName() + " because it already exists")
	}
	cp.Controllers = append(cp.Controllers, ctrl)
	return nil
}

func (cp *RemoteControlPlane) DeleteController(name string) error {
	for idx := range cp.Controllers {
		if cp.Controllers[idx].GetName() == name {
			cp.Controllers = append(cp.Controllers[:idx-1], cp.Controllers[idx+1:]...)
			return nil
		}
	}
	return util.NewError("Could not find Controller " + name + " when performing deletion")
}
