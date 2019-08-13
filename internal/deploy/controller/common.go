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

package deploycontroller

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func prepareUserAndSaveConfig(namespace string, spec config.Controller) (configEntry config.Controller, user config.IofogUser, err error) {
	// Check for existing user
	controlPlane, err := config.GetControlPlane(namespace)
	if err != nil {
		return
	}

	// Return and verify user
	user = controlPlane.IofogUser
	if user.Email == "" || user.Name == "" || user.Password == "" || user.Surname == "" {
		err = util.NewError("Cannot deploy Controller because Control Plane does not have a valid ioFog User")
		return
	}

	spec.Created = util.NowUTC()
	if err = config.UpdateController(namespace, spec); err != nil {
		return
	}
	if err = config.Flush(); err != nil {
		return
	}

	// Return the updated spec
	configEntry = spec

	return
}
