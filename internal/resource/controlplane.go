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
	IofogUser  IofogUser        `yaml:"iofogUser,omitempty"`
	Controller *LocalController `yaml:"controller,omitempty"`
}

type KubernetesControlPlane struct {
	ControllerPods []*KubernetesController `yaml:"controllerPods,omitempty"`
	Database       Database                `yaml:"database,omitempty"`
	IofogUser      IofogUser               `yaml:"iofogUser,omitempty"`
	Endpoint       string                  `yaml:"endpoint,omitempty"`
	KubeConfig     string                  `yaml:"config,omitempty"`
	Services       Services                `yaml:"services,omitempty"`
	Replicas       Replicas                `yaml:"replicas,omitempty"`
	Images         KubeImages              `yaml:"images,omitempty"`
}

type RemoteControlPlane struct {
	Database    Database            `yaml:"database,omitempty"`
	IofogUser   IofogUser           `yaml:"iofogUser,omitempty"`
	Controllers []*RemoteController `yaml:"controllers,omitempty"`
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

func (cp KubernetesControlPlane) GetUser() IofogUser {
	return cp.IofogUser
}

func (cp KubernetesControlPlane) GetControllers() (controllers []Controller) {
	for _, controller := range cp.ControllerPods {
		controllers = append(controllers, controller)
	}
	return
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

func (cp *KubernetesControlPlane) UpdateController(baseController Controller) error {
	controller, ok := baseController.(*KubernetesController)
	if !ok {
		return util.NewError("Must add Kubernetes Controller to Kubernetes Control Plane")
	}
	for idx := range cp.ControllerPods {
		if cp.ControllerPods[idx].GetName() == controller.GetName() {
			cp.ControllerPods[idx] = controller
			return nil
		}
	}
	cp.ControllerPods = append(cp.ControllerPods, controller)
	return nil
}

func (cp *KubernetesControlPlane) AddController(baseController Controller) error {
	if _, err := cp.GetController(baseController.GetName()); err == nil {
		return util.NewError("Could not add Controller " + baseController.GetName() + " because it already exists")
	}
	controller, ok := baseController.(*KubernetesController)
	if !ok {
		return util.NewError("Must add Kubernetes Controller to Kubernetes Control Plane")
	}
	cp.ControllerPods = append(cp.ControllerPods, controller)
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

func (cp RemoteControlPlane) GetControllers() (controllers []Controller) {
	for _, controller := range cp.Controllers {
		controllers = append(controllers, controller)
	}
	return
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

func (cp *RemoteControlPlane) UpdateController(baseController Controller) error {
	controller, ok := baseController.(*RemoteController)
	if !ok {
		return util.NewError("Must add Remote Controller to Remote Control Plane")
	}
	for idx := range cp.Controllers {
		if cp.Controllers[idx].GetName() == controller.GetName() {
			cp.Controllers[idx] = controller
			return nil
		}
	}
	cp.Controllers = append(cp.Controllers, controller)
	return nil
}

func (cp *RemoteControlPlane) AddController(baseController Controller) error {
	if _, err := cp.GetController(baseController.GetName()); err == nil {
		return util.NewError("Could not add Controller " + baseController.GetName() + " because it already exists")
	}
	controller, ok := baseController.(*RemoteController)
	if !ok {
		return util.NewError("Must add Remote Controller to Remote Control Plane")
	}

	cp.Controllers = append(cp.Controllers, controller)
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
