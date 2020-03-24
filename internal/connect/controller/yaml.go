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

package connectcontroller

import (
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
	"gopkg.in/yaml.v2"
)

func unmarshallYAML(file []byte) (controller rsc.RemoteController, err error) {
	// Unmarshall the input file
	if err = yaml.UnmarshalStrict(file, &controller); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}

	// Fix SSH port
	if controller.Host != "" && controller.SSH.Port == 0 {
		controller.SSH.Port = 22
	}
	// Format file paths
	if controller.SSH.KeyFile, err = util.FormatPath(controller.SSH.KeyFile); err != nil {
		return
	}

	return
}

func Validate(ctrl rsc.Controller) error {
	if ctrl.GetName() == "" {
		return util.NewInputError("You must specify a non-empty value for name value of Controllers")
	}
	return nil
}
