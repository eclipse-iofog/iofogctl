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

type Options struct {
	Namespace string
	InputFile string
}

func Execute(opt Options) error {
	_, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return err
	}

	// Unmarshall file
	ctrl, err := UnmarshallYAML(opt.InputFile)
	if err != nil {
		return err
	}

	// Output message
	util.SpinStart("Deploying Controller")

	// Get the Control Plane
	controlPlane, err := config.GetControlPlane(opt.Namespace)
	if err != nil {
		return err
	}

	// Instantiate executor
	exe, err := NewExecutor(opt.Namespace, &ctrl, controlPlane)
	if err != nil {
		return err
	}

	// Execute command
	if err := exe.Execute(); err != nil {
		return err
	}

	// Update configuration
	if err = config.UpdateController(opt.Namespace, ctrl); err != nil {
		return err
	}

	return config.Flush()
}
