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

package deployremotecontrolplane

import (
	"github.com/eclipse-iofog/iofogctl/v2/internal/deploy/controller/remote"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
	"gopkg.in/yaml.v2"
)

// TODO: Unmarshall based on kind?
func UnmarshallYAML(file []byte) (controlPlane *rsc.RemoteControlPlane, err error) {
	// Unmarshall the input file
	if err = yaml.UnmarshalStrict(file, controlPlane); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}
	// None specified
	controllers := controlPlane.GetControllers()
	if len(controlPlane.GetControllers()) == 0 {
		return
	}

	// Validate inputs
	if err = validate(controlPlane); err != nil {
		return
	}

	// Pre-process inputs for Controllers
	for idx := range controllers {
		controller, ok := controllers[idx].(*rsc.RemoteController)
		if !ok {
			err = util.NewInternalError("Could not convert Controller to Remote Controller")
			return
		}
		// Fix SSH port
		if controller.Host != "" && controller.SSH.Port == 0 {
			controller.SSH.Port = 22
		}
		// Format file paths
		if controller.SSH.KeyFile, err = util.FormatPath(controller.SSH.KeyFile); err != nil {
			return
		}
	}

	return
}

func validate(controlPlane *rsc.RemoteControlPlane) (err error) {
	// Validate user
	user := controlPlane.IofogUser
	if user.Email == "" || user.Name == "" || user.Password == "" || user.Surname == "" {
		return util.NewInputError("Control Plane Iofog User must contain non-empty values in email, name, surname, and password fields")
	}
	// Validate database
	db := controlPlane.Database
	if db.Host != "" || db.DatabaseName != "" || db.Password != "" || db.Port != 0 || db.User != "" {
		if db.Host == "" || db.DatabaseName == "" || db.Password == "" || db.Port == 0 || db.User == "" {
			return util.NewInputError("If you are specifying an external database for the Control Plane, you must provide non-empty values in host, databasename, user, password, and port fields,")
		}
	}
	// Validate Controllers
	controllers := controlPlane.GetControllers()
	if len(controllers) == 0 {
		return util.NewInputError("Control Plane must have at least one Controller instance specified.")
	}
	for _, ctrl := range controllers {
		if err = deployremotecontroller.Validate(ctrl); err != nil {
			return
		}
	}

	return
}
