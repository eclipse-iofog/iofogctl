/*
 *  *******************************************************************************
 *  * Copyright (c) 2020 Edgeworx, Inc.
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package deleteremotecontrolplane

import (
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	deletecontroller "github.com/eclipse-iofog/iofogctl/v2/internal/delete/controller"
	"github.com/eclipse-iofog/iofogctl/v2/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type Executor struct {
	namespace string
}

func NewExecutor(namespace string, soft bool) (execute.Executor, error) {
	exe := &Executor{
		namespace: namespace,
	}
	if soft {
		return nil, util.NewInputError("Cannot soft delete a ControlPlane")
	}
	return exe, nil
}

// GetName returns application name
func (exe *Executor) GetName() string {
	return "Delete Control Plane"
}

// Execute deletes application by deleting its associated flow
func (exe *Executor) Execute() (err error) {
	// Get Control Plane
	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	baseControlPlane, err := ns.GetControlPlane()
	if err != nil {
		return err
	}
	controlPlane, ok := baseControlPlane.(*rsc.RemoteControlPlane)
	if !ok {
		return util.NewError("Could not Convert Controller to Remote Controller")
	}

	var executors []execute.Executor
	for _, controller := range controlPlane.GetControllers() {
		exe := deletecontroller.NewRemoteExecutor(controlPlane, exe.namespace, controller.GetName())
		executors = append(executors, exe)
	}

	if err = runExecutors(executors); err != nil {
		return err
	}

	// Delete Control Plane in config
	ns.DeleteControlPlane()
	config.UpdateNamespace(ns)

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
