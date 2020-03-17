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
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

// GetControllers returns all controllers within the namespace
func GetControllers(namespace string) ([]Controller, error) {
	ns, err := getNamespace(namespace)
	if err != nil {
		return nil, err
	}
	return ns.ControlPlane.Controllers, nil
}

// GetController returns a single controller within the current
func GetController(namespace, name string) (controller Controller, err error) {
	ns, err := getNamespace(namespace)
	if err != nil {
		return
	}
	for _, ctrl := range ns.ControlPlane.Controllers {
		if ctrl.Name == name {
			controller = ctrl
			return
		}
	}

	err = util.NewNotFoundError(namespace + "/" + name)
	return
}

// Overwrites or creates new controller to the namespace
func UpdateController(namespace string, controller Controller) error {
	// Update existing controller if exists
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	for idx := range ns.ControlPlane.Controllers {
		if ns.ControlPlane.Controllers[idx].Name == controller.Name {
			mux.Lock()
			ns.ControlPlane.Controllers[idx] = controller
			mux.Unlock()
			return nil
		}
	}
	// Add new controller
	return AddController(namespace, controller)
}

// AddController adds a new controller to the current namespace
func AddController(namespace string, controller Controller) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	_, err = GetController(namespace, controller.Name)
	if err == nil {
		return util.NewConflictError(namespace + "/" + controller.Name)
	}

	// Append the controller
	mux.Lock()
	ns.ControlPlane.Controllers = append(ns.ControlPlane.Controllers, controller)
	mux.Unlock()

	return nil
}

// DeleteController deletes a controller from a namespace
func DeleteController(namespace, name string) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	for idx := range ns.ControlPlane.Controllers {
		if ns.ControlPlane.Controllers[idx].Name == name {
			mux.Lock()
			ns.ControlPlane.Controllers = append(ns.ControlPlane.Controllers[:idx], ns.ControlPlane.Controllers[idx+1:]...)
			mux.Unlock()
			return nil
		}
	}

	return util.NewNotFoundError(ns.Name + "/" + name)
}
