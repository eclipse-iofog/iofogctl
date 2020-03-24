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

package deployk8scontrolplane

import (
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
	"gopkg.in/yaml.v2"
)

func UnmarshallYAML(file []byte) (controlPlane rsc.KubernetesControlPlane, err error) {
	// Unmarshall the input file
	if err = yaml.UnmarshalStrict(file, &controlPlane); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}

	// Validate inputs
	if err = validate(&controlPlane); err != nil {
		return
	}

	// Preprocess Inputs for Control Plane
	if controlPlane.KubeConfig, err = util.FormatPath(controlPlane.KubeConfig); err != nil {
		return
	}
	if controlPlane.Replicas.Controller == 0 {
		controlPlane.Replicas.Controller = 1
	}

	return
}

func validate(controlPlane *rsc.KubernetesControlPlane) (err error) {
	// Validate user
	user := controlPlane.GetUser()
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
	return
}
