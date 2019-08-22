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
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func UnmarshallYAML(filename string) (ctrl config.Controller, err error) {
	// Unmarshall the input file
	if err = util.UnmarshalYAML(filename, &ctrl); err != nil {
		return
	}
	// None specified
	if ctrl.Name == "" || (ctrl.KubeConfig == "" && (ctrl.Host == "" || ctrl.User == "" || ctrl.KeyFile == "")) {
		err = util.NewInputError("Could not unmarshall " + filename + "\n" + err.Error())
		return
	}

	// Fix SSH port
	if ctrl.Port == 0 {
		ctrl.Port = 22
	}
	// Format file paths
	if ctrl.KeyFile, err = util.FormatPath(ctrl.KeyFile); err != nil {
		return
	}
	if ctrl.KubeConfig, err = util.FormatPath(ctrl.KubeConfig); err != nil {
		return
	}

	return
}
