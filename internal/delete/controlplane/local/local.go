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

package deletelocalcontrolplane

import (
	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	deletecontroller "github.com/eclipse-iofog/iofogctl/v3/internal/delete/controller"
	"github.com/eclipse-iofog/iofogctl/v3/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

type Executor struct {
	namespace string
}

func NewExecutor(namespace string) (execute.Executor, error) {
	exe := &Executor{
		namespace: namespace,
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

	controlPlane, ok := baseControlPlane.(*rsc.LocalControlPlane)
	if !ok {
		return util.NewError("Could not convert Control Plane to Local Control Plane")
	}

	executor := deletecontroller.NewLocalExecutor(controlPlane, exe.namespace, controlPlane.Controller.GetName())
	if err := executor.Execute(); err != nil {
		return err
	}

	// Delete Control Plane in config
	ns.DeleteControlPlane()
	return config.Flush()
}
