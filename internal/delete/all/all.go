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

package deleteall

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/delete/agent"
	"github.com/eclipse-iofog/iofogctl/internal/delete/connector"
	"github.com/eclipse-iofog/iofogctl/internal/delete/controller"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func Execute(namespace string) error {
	// Get namespace
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return err
	}

	// Delete Agents
	var executors []execute.Executor
	for _, agent := range ns.Agents {
		exe, err := deleteagent.NewExecutor(namespace, agent.Name)
		if err != nil {
			return err
		}
		executors = append(executors, exe)
	}
	if err := runExecutors(executors); err != nil {
		return err
	}
	for _, agent := range ns.Agents {
		if err = config.DeleteAgent(namespace, agent.Name); err != nil {
			return err
		}
	}

	// Delete Connectors
	executors = executors[:0]
	for _, cnct := range ns.Connectors {
		exe, err := deleteconnector.NewExecutor(namespace, cnct.Name)
		if err != nil {
			return err
		}
		executors = append(executors, exe)
	}
	if err := runExecutors(executors); err != nil {
		return err
	}
	for _, cnct := range ns.Connectors {
		if err = config.DeleteConnector(namespace, cnct.Name); err != nil {
			return err
		}
	}

	// Delete Controllers
	executors = executors[:0]
	for _, ctrl := range ns.ControlPlane.Controllers {
		exe, err := deletecontroller.NewExecutor(namespace, ctrl.Name)
		if err != nil {
			return err
		}
		executors = append(executors, exe)
	}
	if err := runExecutors(executors); err != nil {
		return err
	}
	for _, ctrl := range ns.ControlPlane.Controllers {
		if err = config.DeleteController(namespace, ctrl.Name); err != nil {
			return err
		}
	}

	// Delete Control Plane
	if err = config.DeleteControlPlane(namespace); err != nil {
		return err
	}

	return config.Flush()
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
