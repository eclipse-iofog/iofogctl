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

package deploycontroller

import (
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
	"gopkg.in/yaml.v2"
)

func UnmarshallYAML(file []byte) (ctrl config.Controller, err error) {
	// Unmarshall the input file
	if err = yaml.UnmarshalStrict(file, &ctrl); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}

	// Fix SSH port
	if ctrl.Host != "" && ctrl.SSH.Port == 0 {
		ctrl.SSH.Port = 22
	}
	// Format file paths
	if ctrl.SSH.KeyFile, err = util.FormatPath(ctrl.SSH.KeyFile); err != nil {
		return
	}

	return
}

func Validate(ctrl config.Controller) error {
	if ctrl.Name == "" {
		return util.NewInputError("You must specify a non-empty value for name value of Controllers")
	}
	return nil
}
