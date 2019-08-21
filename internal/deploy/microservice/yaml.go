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

package deploymicroservice

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type specification struct {
	Microservices []config.Microservice
}

func UnmarshallYAML(filename string) (microservices []config.Microservice, err error) {
	// Unmarshall the input file
	var spec specification
	if err = util.UnmarshalYAML(filename, &spec); err != nil || len(spec.Microservices) == 0 {
		var msvc config.Microservice
		if err = util.UnmarshalYAML(filename, &msvc); err != nil {
			err = util.NewInputError("Could not unmarshall " + filename)
			return
		}
		// None specified
		if msvc.Name == "" {
			return
		}
		// Append the single app
		microservices = append(microservices, msvc)
	} else {
		// Record multiple app
		microservices = spec.Microservices
	}

	return
}
