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
	connectcontroller "github.com/eclipse-iofog/iofogctl/v2/internal/connect/controller"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
	"gopkg.in/yaml.v2"
)

func unmarshallKubernetesYaml(file []byte) (controlPlane *rsc.KubernetesControlPlane, err error) {
	// Unmarshall the input file
	if err = yaml.UnmarshalStrict(file, controlPlane); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}
	// None specified
	if len(controlPlane.GetControllers()) == 0 {
		err = util.NewInputError("No Controllers specified in Control Plane. Cannot connect.")
		return
	}
	// Preprocess Inputs for Control Plane
	if controlPlane.KubeConfig, err = util.FormatPath(controlPlane.KubeConfig); err != nil {
		return
	}

	// Validate inputs
	if err = validate(controlPlane); err != nil {
		return
	}

	return
}

func unmarshallRemoteYAML(file []byte) (controlPlane *rsc.RemoteControlPlane, err error) {
	// Unmarshall the input file
	if err = yaml.UnmarshalStrict(file, controlPlane); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}
	// None specified
	if len(controlPlane.GetControllers()) == 0 {
		err = util.NewInputError("No Controllers specified in Control Plane. Cannot connect.")
		return
	}
	// Pre-process controllers
	for idx := range controlPlane.GetControllers() {
		if controlPlane.Controllers[idx].SSH.KeyFile, err = util.FormatPath(controlPlane.Controllers[idx].SSH.KeyFile); err != nil {
			return
		}
	}

	controlPlane = controlPlane

	// Validate inputs
	if err = validate(controlPlane); err != nil {
		return
	}

	return
}

func unmarshallYAML(file []byte) (controlPlane rsc.ControlPlane, err error) {
	// Unmarshall the input file
	var controlPlane rsc.ControlPlane
	if err = yaml.UnmarshalStrict(file, &controlPlane); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}
	// None specified
	if len(controlPlane.GetControllers()) == 0 {
		err = util.NewInputError("No Controllers specified in Control Plane. Cannot connect.")
		return
	}
	if controlPlane.Kube.Config, err = util.FormatPath(controlPlane.Kube.Config); err != nil {
		return
	}
	// Preprocess Inputs for Control Plane
	if controlPlane.Kube.Config, err = util.FormatPath(controlPlane.Kube.Config); err != nil {
		return
	}
	// Pre-process controllers
	for idx := range controlPlane.Controllers {
		if controlPlane.Controllers[idx].SSH.KeyFile, err = util.FormatPath(controlPlane.Controllers[idx].SSH.KeyFile); err != nil {
			return
		}
	}

	controlPlane = controlPlane

	// Validate inputs
	if err = validate(controlPlane); err != nil {
		return
	}

	return
}

func validate(controlPlane rsc.ControlPlane) (err error) {
	// Validate user
	user := controlPlane.GetUser()
	if user.Password == "" || user.Email == "" {
		return util.NewInputError("To connect, Control Plane Iofog User must contain non-empty values in email and password fields")
	}
	// Validate Controllers
	if len(controlPlane.Controllers) == 0 {
		return util.NewInputError("Control Plane must have at least one Controller instance specified.")
	}
	for _, ctrl := range controlPlane.GetControllers() {
		if err = connectcontroller.Validate(ctrl); err != nil {
			return
		}
	}

	return
}
