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

package deployagent

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

	// Read the input file
	agents, err := UnmarshallYAML(opt.InputFile)
	if err != nil {
		return err
	}

	// Output message
	msg := "Deploying Agent"
	if len(agents) > 1 {
		msg += "s"
	}
	util.SpinStart(msg)

	// Instantiate executors
	var executors []execute.Executor
	for idx := range agents {
		exe, err := newExecutor(ns.Name, &agents[idx])
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

	// Update configuration
	for idx := range agents {
		if err = config.UpdateAgent(opt.Namespace, agents[idx]); err != nil {
			return err
		}
	}

	return config.Flush()
}
