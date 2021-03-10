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

package controller

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

func Execute(namespace, name, newName string) error {
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return err
	}
	// Check that Controller exists in current namespace
	controlPlane, err := ns.GetControlPlane()
	if err != nil {
		return err
	}

	// Get the Controller to rename
	controller, err := controlPlane.GetController(name)
	if err != nil {
		return err
	}

	// Check new name is valid
	if err := util.IsLowerAlphanumeric("Controller", newName); err != nil {
		return err
	}

	// Perform the rename
	util.SpinStart(fmt.Sprintf("Renaming Controller %s", name))
	controller.SetName(newName)
	ns.SetControlPlane(controlPlane)

	return config.Flush()
}
