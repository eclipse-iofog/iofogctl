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

package deletecontroller

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func NewExecutor(namespace, name string) (execute.Executor, error) {
	// Get controller from config
	ctrl, err := config.GetController(namespace, name)
	if err != nil {
		return nil, err
	}

	// Local executor
	if util.IsLocalHost(ctrl.Host) {
		cli, err := install.NewLocalContainerClient()
		if err != nil {
			return nil, err
		}
		return newLocalExecutor(namespace, name, cli), nil
	}

	// Kubernetes executor
	if ctrl.Kube.Config != "" {
		return newKubernetesExecutor(namespace, name), nil
	}

	// Can't kill Controller without configuration
	if ctrl.Host == "" || ctrl.SSH.User == "" || ctrl.SSH.KeyFile == "" || ctrl.SSH.Port == 0 {
		return nil, util.NewNoConfigError("Controller")
	}
	// Default executor
	return newRemoteExecutor(namespace, name), nil
}
