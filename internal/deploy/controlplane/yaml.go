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

package deploycontrolplane

import (
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	deploycontroller "github.com/eclipse-iofog/iofogctl/v2/internal/deploy/controller"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
	"gopkg.in/yaml.v2"
)

func UnmarshallYAML(file []byte) (controlPlane rsc.ControlPlane, err error) {
	// Unmarshall the input file
	var ctrlPlane rsc.ControlPlane
	if err = yaml.UnmarshalStrict(file, &ctrlPlane); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}
	// None specified
	if len(ctrlPlane.Controllers) == 0 {
		return
	}
	controlPlane = ctrlPlane

	// Validate inputs
	if err = validate(controlPlane); err != nil {
		return
	}

	// Preprocess Inputs for Control Plane
	if controlPlane.Kube.Config, err = util.FormatPath(controlPlane.Kube.Config); err != nil {
		return
	}

	// Pre-process inputs for Controllers
	for idx := range controlPlane.Controllers {
		ctrl := &controlPlane.Controllers[idx]
		// Fix SSH port
		if ctrl.Host != "" && ctrl.SSH.Port == 0 {
			ctrl.SSH.Port = 22
		}
		// Format file paths
		if ctrl.SSH.KeyFile, err = util.FormatPath(ctrl.SSH.KeyFile); err != nil {
			return
		}
	}

	return
}

func validate(controlPlane rsc.ControlPlane) (err error) {
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
	// Validate loadbalancer
	lb := controlPlane.LoadBalancer
	if lb.Host != "" || lb.Port != 0 {
		if lb.Host == "" || lb.Port == 0 {
			return util.NewInputError("If you are specifying a load balancer you must provide non-empty valies in host and port fields")
		}
	}
	// Validate Controllers
	if len(controlPlane.Controllers) == 0 {
		return util.NewInputError("Control Plane must have at least one Controller instance specified.")
	}
	for _, ctrl := range controlPlane.Controllers {
		if err = deploycontroller.Validate(ctrl); err != nil {
			return
		}
	}

	return
}
