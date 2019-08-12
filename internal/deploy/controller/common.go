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

func prepareUserAndSaveConfig(namespace string, spec config.Controller) (configEntry config.Controller, err error) {
	var configUser config.IofogUser
	// Check for existing user
	ctrl, err := config.GetController(namespace, spec.Name)
	if spec.IofogUser.Email != "" && spec.IofogUser.Password != "" {
		// Use user provided in the yaml file
		configUser = spec.IofogUser
	} else if err == nil {
		// Use existing user
		configUser = ctrl.IofogUser
	} else {
		// Generate new user
		configUser = config.NewRandomUser()
	}

	// Record the user
	spec.IofogUser = configUser
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
