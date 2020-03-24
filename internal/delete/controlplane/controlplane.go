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

package deletecontrolplane

import (
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	deletecontroller "github.com/eclipse-iofog/iofogctl/v2/internal/delete/controller"
	"github.com/eclipse-iofog/iofogctl/v2/internal/execute"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type Executor struct {
	namespace string
	name      string
}

func NewExecutor(namespace, name string, soft bool) (execute.Executor, error) {
	exe := &Executor{
		namespace: namespace,
		name:      name,
	}
	if soft {
		return nil, util.NewInputError("Cannot soft delete a ControlPlane")
	}
	return exe, nil
}

// GetName returns application name
func (exe *Executor) GetName() string {
	return exe.name
}

// Execute deletes application by deleting its associated flow
func (exe *Executor) Execute() (err error) {
	// Get Control Plane
	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	controlPlane, err := ns.GetControlPlane()
	if err != nil {
		return err
	}

	var executors []execute.Executor
	for _, controller := range controlPlane.GetControllers() {
		exe, err := deletecontroller.NewExecutor(exe.namespace, controller.GetName(), false)
		if err != nil {
			return err
		}
		executors = append(executors, exe)
	}

	if err = runExecutors(executors); err != nil {
		return err
	}

	// Delete Control Plane
	controlPlane = nil

	return config.Flush()
}

func runExecutors(executors []execute.Executor) error {
	if errs, failedExes := execute.ForParallel(executors); len(errs) > 0 {
		for idx := range errs {
			util.PrintNotify("Error from " + failedExes[idx].GetName() + ": " + errs[idx].Error())
		}
		return util.NewError("Failed to delete")
	}
	return nil
}
