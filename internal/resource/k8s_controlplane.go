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

type KubernetesControlPlane struct {
	KubeConfig     string                 `yaml:"config"`
	IofogUser      IofogUser              `yaml:"iofogUser"`
	ControllerPods []KubernetesController `yaml:"controllerPods,omitempty"`
	Database       Database               `yaml:"database,omitempty"`
	Services       Services               `yaml:"services,omitempty"`
	Replicas       Replicas               `yaml:"replicas,omitempty"`
	Images         KubeImages             `yaml:"images,omitempty"`
	Endpoint       string                 `yaml:"endpoint,omitempty"`
}

func (cp KubernetesControlPlane) GetUser() IofogUser {
	return cp.IofogUser
}

func (cp KubernetesControlPlane) GetControllers() (controllers []Controller) {
	for idx := range cp.ControllerPods {
		controllers = append(controllers, &cp.ControllerPods[idx])
	}
	return
}

func (cp KubernetesControlPlane) GetController(name string) (ret Controller, err error) {
	for _, ctrl := range cp.ControllerPods {
		if ctrl.GetName() == name {
			ret = &ctrl
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
			cp.ControllerPods[idx] = *controller
			return nil
		}
	}
	cp.ControllerPods = append(cp.ControllerPods, *controller)
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
	cp.ControllerPods = append(cp.ControllerPods, *controller)
	return nil
}

func (cp *KubernetesControlPlane) DeleteController(name string) error {
	for idx := range cp.ControllerPods {
		if cp.ControllerPods[idx].GetName() == name {
			cp.ControllerPods = append(cp.ControllerPods[:idx], cp.ControllerPods[idx+1:]...)
			return nil
		}
	}
	return util.NewError("Could not find Controller " + name + " when performing deletion")
}

func (cp *KubernetesControlPlane) Sanitize() (err error) {
	if cp.KubeConfig, err = util.FormatPath(cp.KubeConfig); err != nil {
		return
	}
	if cp.Replicas.Controller == 0 {
		cp.Replicas.Controller = 1
	}
	return
}
