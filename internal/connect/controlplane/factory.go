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

package connectcontrolplane

import (
	deploycontrolplane "github.com/eclipse-iofog/iofogctl/internal/deploy/controlplane"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
)

func NewExecutor(namespace, name string, yaml []byte) (execute.Executor, error) {
	// Read the input file
	controlPlane, err := deploycontrolplane.UnmarshallYAML(yaml)
	if err != nil {
		return nil, err
	}

	// Kubernetes controller
	if controlPlane.Controllers[0].KubeConfig != "" {
		return newKubernetesExecutor(controlPlane, namespace), nil
	}

	return newRemoteExecutor(controlPlane, namespace), nil
}
