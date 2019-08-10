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
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
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

	// Read the input file
	agents, err := UnmarshallYAML(opt.InputFile)
	if err != nil {
		return err
	}

	// Instantiate executors
	var executors []execute.Executor
	for idx := range agents {
		exe, err := newExecutor(ns.Name, agents[idx])
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

	return nil
}

func newExecutor(namespace string, agent config.Agent) (execute.Executor, error) {
	// Check the namespace exists
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return nil, err
	}

	// Check Controller exists
	if len(ns.ControlPlane.Controllers) == 0 {
		return nil, util.NewInputError("This namespace does not have a Controller. You must first deploy a Controller before deploying Agents")
	}

	// Local executor
	if util.IsLocalHost(agent.Host) {
		cli, err := install.NewLocalContainerClient()
		if err != nil {
			return nil, err
		}
		exe, err := newLocalExecutor(namespace, agent, cli)
		if err != nil {
			return nil, err
		}
		return exe, nil
	}

	// Default executor
	if agent.Host == "" || agent.KeyFile == "" || agent.User == "" {
		return nil, util.NewInputError("Must specify user, host, and key file flags for remote deployment")
	}
	return newRemoteExecutor(namespace, agent), nil
}
