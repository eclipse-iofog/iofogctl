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

package connectcontrolplane

import (
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	connectcontroller "github.com/eclipse-iofog/iofogctl/v2/internal/connect/controller"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
	"gopkg.in/yaml.v2"
)

func unmarshallYAML(file []byte) (controlPlane config.ControlPlane, err error) {
	// Unmarshall the input file
	var ctrlPlane config.ControlPlane
	if err = yaml.UnmarshalStrict(file, &ctrlPlane); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}
	// None specified
	if len(ctrlPlane.Controllers) == 0 {
		err = util.NewInputError("No Controllers specified in Control Plane. Cannot connect.")
		return
	}
	if ctrlPlane.Kube.Config, err = util.FormatPath(ctrlPlane.Kube.Config); err != nil {
		return
	}
	// Preprocess Inputs for Control Plane
	if ctrlPlane.Kube.Config, err = util.FormatPath(ctrlPlane.Kube.Config); err != nil {
		return
	}
	// Pre-process controllers
	for idx := range ctrlPlane.Controllers {
		if ctrlPlane.Controllers[idx].SSH.KeyFile, err = util.FormatPath(ctrlPlane.Controllers[idx].SSH.KeyFile); err != nil {
			return
		}
	}

	controlPlane = ctrlPlane

	// Validate inputs
	if err = validate(controlPlane); err != nil {
		return
	}

	return
}

func validate(controlPlane config.ControlPlane) (err error) {
	// Validate user
	user := controlPlane.IofogUser
	if user.Password == "" || user.Email == "" {
		return util.NewInputError("To connect, Control Plane Iofog User must contain non-empty values in email and password fields")
	}
	// Validate Controllers
	if len(controlPlane.Controllers) == 0 {
		return util.NewInputError("Control Plane must have at least one Controller instance specified.")
	}
	for _, ctrl := range controlPlane.Controllers {
		if err = connectcontroller.Validate(ctrl); err != nil {
			return
		}
	}

	return
}
