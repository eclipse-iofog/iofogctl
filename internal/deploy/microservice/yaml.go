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
	deploy "github.com/eclipse-iofog/iofog-go-sdk/pkg/deployapps"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"gopkg.in/yaml.v2"
)

type specification struct {
	Microservices []deploy.Microservice
}

func UnmarshallYAML(file []byte) (microservices []deploy.Microservice, err error) {
	// Unmarshall the input file
	var spec specification
	if err = yaml.Unmarshal(file, &spec); err != nil || len(spec.Microservices) == 0 {
		var msvc deploy.Microservice
		if err = yaml.Unmarshal(file, &msvc); err != nil {
			err = util.NewInputError("Could not unmarshall\n" + err.Error())
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
