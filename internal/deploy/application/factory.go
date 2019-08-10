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

package deployapplication

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Options struct {
	Namespace string
	InputFile string
}

type jobResult struct {
	name string
	err  error
}

func Deploy(opt Options) error {
	// Check the namespace exists
	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return err
	}

	// Check Controller exists
	nbControllers := len(ns.ControlPlane.Controllers)
	if nbControllers != 1 {
		errMessage := fmt.Sprintf("This namespace contains %d Controller(s), you must have one, and only one.", nbControllers)
		return util.NewInputError(errMessage)
	}

	applications, err := UnmarshallYAML(opt.InputFile)
	if err != nil {
		return err
	}

	// Instantiate executors
	var executors []execute.Executor
	for idx := range applications {
		exe, err := newExecutor(ns.Name, applications[idx])
		if err != nil {
			return err
		}
		executors = append(executors, exe)
	}
	// Execute
	if err = execute.ForParallel(executors); err != nil {
		return err
	}

	return nil
}

func newExecutor(namespace string, opt config.Application) (execute.Executor, error) {
	return newRemoteExecutor(namespace, opt), nil
}
