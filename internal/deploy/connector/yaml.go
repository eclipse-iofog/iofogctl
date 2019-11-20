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
	"gopkg.in/yaml.v2"
)

func UnmarshallYAML(file []byte) (connector config.Connector, err error) {
	// Unmarshall the input file
	if err = yaml.UnmarshalStrict(file, &connector); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}
	// None specified
	if connector.Kube.Config == "" && (connector.Host == "" || connector.SSH.User == "" || connector.SSH.KeyFile == "") {
		return
	}

	// Pre-process the fields
	// Fix SSH port
	if connector.Host != "" && connector.SSH.Port == 0 {
		connector.SSH.Port = 22
	}
	// Format file paths
	if connector.SSH.KeyFile, err = util.FormatPath(connector.SSH.KeyFile); err != nil {
		return
	}
	if connector.Kube.Config, err = util.FormatPath(connector.Kube.Config); err != nil {
		return
	}

	return
}

func Validate(cnct config.Connector) error {
	if cnct.Name == "" {
		return util.NewInputError("You must specify a non-empty value for name value of Connectors")
	}
	if cnct.Kube.Config == "" && ((cnct.Host != "localhost" && cnct.Host != "127.0.0.1") && (cnct.Host == "" || cnct.SSH.User == "" || cnct.SSH.KeyFile == "")) {
		return util.NewInputError("For Connectors you must specify non-empty values for host, user, and keyfile")
	}
	return nil
}
