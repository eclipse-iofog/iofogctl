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
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

// GetControllers returns all controllers within the namespace
func GetControllers(namespace string) ([]rsc.Controller, error) {
	ns, err := getNamespace(namespace)
	if err != nil {
		return nil, err
	}
	return ns.ControlPlane.GetControllers(), nil
}

// GetController returns a single controller within the current
func GetController(namespace, name string) (controller rsc.Controller, err error) {
	ns, err := getNamespace(namespace)
	if err != nil {
		return
	}
	for _, ctrl := range ns.ControlPlane.GetControllers() {
		if ctrl.GetName() == name {
			controller = ctrl
			return
		}
	}

	err = util.NewNotFoundError(namespace + "/" + name)
	return
}

// Overwrites or creates new controller to the namespace
func UpdateController(namespace string, controller rsc.Controller) error {
	// Update existing controller if exists
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	mux.Lock()
	if err := ns.ControlPlane.UpdateController(controller); err != nil {
		return err
	}
	mux.Unlock()
	return nil
}

// AddController adds a new controller to the current namespace
func AddController(namespace string, controller rsc.Controller) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	_, err = ns.ControlPlane.GetController(controller.GetName())
	if err == nil {
		return util.NewConflictError(namespace + "/" + controller.GetName())
	}

	// Add the Controller
	mux.Lock()
	if err := ns.ControlPlane.AddController(controller); err != nil {
		return err
	}
	mux.Unlock()

	return nil
}

// DeleteController deletes a controller from a namespace
func DeleteController(namespace, name string) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}

	mux.Lock()
	if err := ns.ControlPlane.DeleteController(name); err != nil {
		return err
	}
	mux.Unlock()
	return nil
}
