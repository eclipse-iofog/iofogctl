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
	"github.com/eclipse-iofog/iofogctl/v2/internal/execute"
)

type Options struct {
	Namespace string
	Yaml      []byte
	Name      string
}

func NewExecutorWithoutParsing(namespace string, controller *config.Controller, controlPlane config.ControlPlane) (exe execute.Executor, err error) {
	_, err = config.GetNamespace(namespace)
	if err != nil {
		return
	}

	// Instantiate executor
	return newExecutor(namespace, controller, controlPlane)
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	// Unmarshall file
	ctrl, err := UnmarshallYAML(opt.Yaml)
	if err != nil {
		return
	}

	if len(opt.Name) > 0 {
		ctrl.Name = opt.Name
	}

	// Validate
	if err = Validate(ctrl); err != nil {
		return
	}

	// Get the Control Plane
	controlPlane, err := config.GetControlPlane(opt.Namespace)
	if err != nil {
		return
	}

	return NewExecutorWithoutParsing(opt.Namespace, &ctrl, controlPlane)
}
