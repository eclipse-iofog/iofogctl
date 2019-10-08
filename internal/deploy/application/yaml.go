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

package deployapplication

import (
	deploy "github.com/eclipse-iofog/iofog-go-sdk/pkg/deployapps"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"gopkg.in/yaml.v2"
)

type specification struct {
	Applications []deploy.Application
}

func UnmarshallYAML(file []byte) (applications []deploy.Application, err error) {
	// Unmarshall the input file
	var spec specification
	if err = yaml.Unmarshal(file, &spec); err != nil || len(spec.Applications) == 0 {
		var app deploy.Application
		if err = yaml.Unmarshal(file, &app); err != nil {
			err = util.NewInputError("Could not unmarshall\n" + err.Error())
			return
		}
		// None specified
		if app.Name == "" {
			return
		}
		// Append the single app
		applications = append(applications, app)
	} else {
		// Record multiple app
		applications = spec.Applications
	}

	return
}
