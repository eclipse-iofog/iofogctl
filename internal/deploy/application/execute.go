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
	"github.com/eclipse-iofog/iofogctl/internal/config"
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

	// Check Controller exists
	if len(ns.ControlPlane.Controllers) == 0 {
		return util.NewInputError("This namespace does not have a Controller. You must first deploy a Controller before deploying Applications")
	}

	// Unmarshal file
	applications, err := UnmarshallYAML(opt.InputFile)
	if err != nil {
		return err
	}

	// Should be atleast one to deploy
	if len(applications) == 0 {
		return util.NewError("Could not read any Applications from YAML")
	}

	// Output message
	msg := "Deploying Application"
	if len(applications) > 1 {
		msg += "s"
	}
	util.SpinStart(msg)

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
	if errs, failedExes := execute.ForParallel(executors); len(errs) > 0 {
		for idx := range errs {
			util.PrintNotify("Error from " + failedExes[idx].GetName() + ": " + errs[idx].Error())
		}
		return util.NewError("Failed to deploy")
	}

	return config.Flush()
}
