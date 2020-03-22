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

package config

import (
	"strconv"

	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type ControlPlane interface {
	GetControllers() []Controller
	GetController(string) Controller
	GetEndpoint() string
}

type KubernetesControlPlane struct {
	Database     Database     `yaml:"database,omitempty"`
	LoadBalancer LoadBalancer `yaml:"loadBalancer,omitempty"`
	IofogUser    IofogUser    `yaml:"iofogUser,omitempty"`
	KubeConfig   string       `yaml:"config,omitempty"`
	Services     Services     `yaml:"services,omitempty"`
	Replicas     Replicas     `yaml:"replicas,omitempty"`
	Images       KubeImages   `yaml:"images,omitempty"`
}

type LocalControlPlane struct {
	IofogUser IofogUser `yaml:"iofogUser,omitempty"`
}

type RemoteControlPlane struct {
	Database    Database     `yaml:"database,omitempty"`
	IofogUser   IofogUser    `yaml:"iofogUser,omitempty"`
	Controllers []Controller `yaml:"controllers,omitempty"`
}

func (ctrlPlane KubernetesControlPlane) GetController(name string) (ctrl Controller, err error) {
	if len(ctrlPlane.Controllers) == 0 {
		err = util.NewError("Control Plane has no Controllers")
	}
	for _, controller := range ctrlPlane.Controllers {
		if controller.Name == name {
			ctrl = controller
			return
		}
	}

	err = util.NewNotFoundError(name)
	return
}

// GetControllerEndpoint returns ioFog controller endpoint
func (ctrlPlane ControlPlane) GetControllerEndpoint() (string, error) {
	// Loadbalancer ?
	if ctrlPlane.LoadBalancer.Host != "" {
		if ctrlPlane.LoadBalancer.Port != 0 {
			return ctrlPlane.LoadBalancer.Host + ":" + strconv.Itoa(ctrlPlane.LoadBalancer.Port), nil
		}
		return ctrlPlane.LoadBalancer.Host + ":" + iofog.ControllerPortString, nil
	}

	// First controller
	if len(ctrlPlane.Controllers) < 1 {
		return "", util.NewError("This control plane does not have controller")
	}
	return ctrlPlane.Controllers[0].Endpoint, nil
}

func DeleteControlPlane(namespace string) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	mux.Lock()
	ns.ControlPlane = ControlPlane{}
	mux.Unlock()
	return nil
}

// GetControlPlane returns a control plane within a namespace
func GetControlPlane(namespace string) (ControlPlane, error) {
	ns, err := getNamespace(namespace)
	if err != nil {
		return ControlPlane{}, err
	}
	return ns.ControlPlane, nil
}

// UpdateControlPlane overwrites Control Plane in the namespace
func UpdateControlPlane(namespace string, controlPlane ControlPlane) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	mux.Lock()
	ns.ControlPlane = controlPlane
	mux.Unlock()
	return nil
}
