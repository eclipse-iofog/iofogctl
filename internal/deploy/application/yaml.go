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
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type specification struct {
	Applications []config.Application
}

func UnmarshallYAML(filename string) (applications []config.Application, err error) {
	// Unmarshall the input file
	var spec specification
	if err = util.UnmarshalYAML(filename, &spec); err != nil || len(spec.Applications) == 0 {
		var app config.Application
		if err = util.UnmarshalYAML(filename, &app); err != nil {
			err = util.NewInputError("Could not unmarshall " + filename)
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
