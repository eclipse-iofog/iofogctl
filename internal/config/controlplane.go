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
)

func DeleteControlPlane(namespace string) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	mux.Lock()
	ns.ControlPlane = nil
	mux.Unlock()
	return nil
}

// GetControlPlane returns a control plane within a namespace
func GetControlPlane(namespace string) (rsc.ControlPlane, error) {
	ns, err := getNamespace(namespace)
	if err != nil {
		return nil, err
	}
	return ns.ControlPlane, nil
}

// UpdateControlPlane overwrites Control Plane in the namespace
func UpdateControlPlane(namespace string, controlPlane rsc.ControlPlane) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	mux.Lock()
	ns.ControlPlane = controlPlane
	mux.Unlock()
	return nil
}
