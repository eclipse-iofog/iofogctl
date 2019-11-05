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

package connectconnector

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"gopkg.in/yaml.v2"
)

func unmarshallYAML(file []byte) (connector config.Connector, err error) {
	// Unmarshall the input file
	if err = yaml.UnmarshalStrict(file, &connector); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}

	// Pre-process the fields
	// Fix SSH port
	if connector.SSH.Port == 0 {
		connector.SSH.Port = 22
	}
	if connector.KubeConfig, err = util.FormatPath(connector.KubeConfig); err != nil {
		return
	}

	return
}
