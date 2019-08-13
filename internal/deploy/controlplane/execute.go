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
	"github.com/eclipse-iofog/iofogctl/internal/deploy/connector"
	"github.com/eclipse-iofog/iofogctl/internal/deploy/controller"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Options struct {
	Namespace string
	InputFile string
}

func Execute(opt Options) error {
	// Check the namespace exists
	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return err
	}

	// Read the input file
	spec, err := UnmarshallYAML(opt.InputFile)
	if err != nil {
		return err
	}

	// Generate spec.IofogUser if required
	if spec.IofogUser.Email == "" || spec.IofogUser.Name == "" || spec.IofogUser.Password == "" || spec.IofogUser.Surname == "" {
		util.PrintNotify("Generating random ioFog spec.IofogUser because name, surname, email, and password were not supplied")
		spec.IofogUser = config.NewRandomUser()
	}

	// Instantiate executors
	var executors []execute.Executor

	// Execute Controllers
	for idx := range spec.Controllers {
		exe, err := deploycontroller.NewExecutor(ns.Name, spec.Controllers[idx])
		if err != nil {
			return err
		}
		executors = append(executors, exe)
	}
	if err := runExecutors(executors); err != nil {
		return err
	}

	// Execute Connectors
	executors = executors[:0]
	for idx := range spec.Connectors {
		exe, err := deployconnector.NewExecutor(ns.Name, spec.Connectors[idx])
		if err != nil {
			return err
		}
		executors = append(executors, exe)
	}
	if err := runExecutors(executors); err != nil {
		return err
	}

	return nil
}

func runExecutors(executors []execute.Executor) error {
	if errs, failedExes := execute.ForParallel(executors); len(errs) > 0 {
		for idx := range errs {
			util.PrintNotify("Error from " + failedExes[idx].GetName() + ": " + errs[idx].Error())
		}
		return util.NewError("Failed to deploy")
	}
	return nil
}
