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
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
)

type Options struct {
	Namespace string
	Yaml      []byte
	Name      string
	Kind      config.Kind
}

func NewExecutorWithoutParsing(namespace string, controller rsc.Controller) (exe execute.Executor, err error) {
	_, err = config.GetNamespace(namespace)
	if err != nil {
		return
	}

	// Instantiate executor
	return newExecutor(namespace, controller)
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	// Unmarshall file
	ctrl, err := UnmarshallYAML(opt.Yaml)
	if err != nil {
		return
	}

	if opt.Name != "" {
		ctrl.SetName(opt.Name)
	}

	// Validate
	if err = Validate(ctrl); err != nil {
		return
	}

	return NewExecutorWithoutParsing(opt.Namespace, ctrl)
}
