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

package deploycontrolplane

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type specification struct {
	ControlPlane config.ControlPlane
}

func UnmarshallYAML(filename string) (controlPlane config.ControlPlane, err error) {
	// Unmarshall the input file
	var spec specification
	if err = util.UnmarshalYAML(filename, &spec); err != nil || len(spec.ControlPlane.Controllers) == 0 {
		var ctrlPlane config.ControlPlane
		if err = util.UnmarshalYAML(filename, &ctrlPlane); err != nil {
			err = util.NewInputError("Could not unmarshall " + filename)
			return
		}
		// None specified
		if len(ctrlPlane.Controllers) == 0 {
			return
		}
		controlPlane = ctrlPlane
	} else {
		controlPlane = spec.ControlPlane
	}

	// Pre-process inputs
	for idx := range controlPlane.Controllers {
		ctrl := &controlPlane.Controllers[idx]
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
	}

	return
}
