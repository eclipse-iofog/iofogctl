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

package deployconnector

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type specification struct {
	Connectors []config.Connector
}

func UnmarshallYAML(filename string) (connectors []config.Connector, err error) {
	// Unmarshall the input file
	var spec specification
	if err = util.UnmarshalYAML(filename, &spec); err != nil || len(spec.Connectors) == 0 {
		var cnct config.Connector
		if err = util.UnmarshalYAML(filename, &cnct); err != nil {
			err = util.NewInputError("Could not unmarshall " + filename + "\n" + err.Error())
			return
		}
		// None specified
		if cnct.Name == "" || (cnct.KubeConfig == "" && (cnct.Host == "" || cnct.User == "" || cnct.KeyFile == "")) {
			return
		}
		// Validate
		if err = validate(cnct); err != nil {
			return
		}

		// Append the single cnct
		connectors = append(connectors, cnct)
	} else {
		// Record multiple cnct
		connectors = spec.Connectors
	}

	// Pre-process the fields
	for idx := range connectors {
		cnct := &connectors[idx]
		// Fix SSH port
		if cnct.Port == 0 {
			cnct.Port = 22
		}
		// Format file paths
		if cnct.KeyFile, err = util.FormatPath(cnct.KeyFile); err != nil {
			return
		}
		if cnct.KubeConfig, err = util.FormatPath(cnct.KubeConfig); err != nil {
			return
		}
	}

	return
}

func validate(cnct config.Connector) error {
	if cnct.Name == "" {
		return util.NewInputError("You must specify a non-empty value for name value of Connectors")
	}
	if cnct.KubeConfig == "" && ((cnct.Host != "localhost" && cnct.Host != "127.0.0.1") && (cnct.Host == "" || cnct.User == "" || cnct.KeyFile == "")) {
		return util.NewInputError("For Connectors you must specify non-empty values for host, user, and keyfile")
	}
	return nil
}
