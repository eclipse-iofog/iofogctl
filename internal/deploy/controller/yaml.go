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
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
	"gopkg.in/yaml.v2"
)

func UnmarshallYAML(file []byte) (ctrl rsc.Controller, err error) {
	// Unmarshall the input file
	if err = yaml.UnmarshalStrict(file, &ctrl); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}

	switch dynamicCtrl := ctrl.(type) {
	case *rsc.RemoteController:
		// Format file paths
		if dynamicCtrl.SSH.KeyFile, err = util.FormatPath(dynamicCtrl.SSH.KeyFile); err != nil {
			return
		}
		// Fix SSH port
		if dynamicCtrl.Host != "" && dynamicCtrl.SSH.Port == 0 {
			dynamicCtrl.SSH.Port = 22
		}
	}

	return
}

func Validate(ctrl rsc.Controller) error {
	if ctrl.GetName() == "" {
		return util.NewInputError("You must specify a non-empty value for name value of Controllers")
	}
	return nil
}
