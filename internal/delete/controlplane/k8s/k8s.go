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

package deletek8scontrolplane

import (
	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	"github.com/eclipse-iofog/iofogctl/v3/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/iofog/install"
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

	controlPlane, ok := baseControlPlane.(*rsc.KubernetesControlPlane)
	if !ok {
		return util.NewError("Could not convert Control Plane to Kubernetes Control Plane")
	}

	// Instantiate Kubernetes object
	k8s, err := install.NewKubernetes(controlPlane.KubeConfig, exe.namespace)
	if err != nil {
		return err
	}

	// Delete Controller on cluster
	err = k8s.DeleteControlPlane()
	if err != nil {
		return err
	}

	// Delete Control Plane in config
	ns.DeleteControlPlane()

	return config.Flush()
}
