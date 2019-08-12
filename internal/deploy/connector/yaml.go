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

func UnmarshallYAML(filename string) (cnct config.Connector, err error) {
	// Unmarshall the input file
	if err = util.UnmarshalYAML(filename, &cnct); err != nil {
		return
	}
	// None specified
	if cnct.Name == "" || cnct.Host == "" || cnct.User == "" || cnct.KeyFile == "" {
		err = util.NewInputError("Could not unmarshal YAML file")
		return
	}

	// Fix SSH port
	if cnct.Port == 0 {
		cnct.Port = 22
	}
	// Format file paths
	if cnct.KeyFile, err = util.FormatPath(cnct.KeyFile); err != nil {
		return
	}

	return
}
