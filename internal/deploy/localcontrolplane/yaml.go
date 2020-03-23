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

package deploylocalcontrolplane

import (
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
	"gopkg.in/yaml.v2"
)

// TODO: Unmarshall based on kind?
func UnmarshallYAML(file []byte) (controlPlane *rsc.LocalControlPlane, err error) {
	// Unmarshall the input file
	if err = yaml.UnmarshalStrict(file, &controlPlane); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}

	// Validate inputs
	if err = validate(controlPlane); err != nil {
		return
	}

	return
}

func validate(controlPlane rsc.ControlPlane) (err error) {
	// Validate user
	user := controlPlane.GetUser()
	if user.Email == "" || user.Name == "" || user.Password == "" || user.Surname == "" {
		return util.NewInputError("Control Plane Iofog User must contain non-empty values in email, name, surname, and password fields")
	}

	return
}
