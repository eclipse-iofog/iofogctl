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

package deleteconnector

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func NewExecutor(namespace, name string) (execute.Executor, error) {
	// Get controller from config
	cnct, err := config.GetConnector(namespace, name)
	if err != nil {
		return nil, err
	}

	// Local executor
	if util.IsLocalHost(cnct.Host) {
		cli, err := install.NewLocalContainerClient()
		if err != nil {
			return nil, err
		}
		return newLocalExecutor(namespace, name, cli), nil
	}

	// Kubernetes executor
	if cnct.KubeConfig != "" {
		return newKubernetesExecutor(namespace, name), nil
	}

	// Default executor
	if cnct.Host == "" || cnct.User == "" || cnct.KeyFile == "" || cnct.Port == 0 {
		return nil, util.NewError("Cannot execute delete command because Kube Config and SSH details for this Connector are not available")
	}
	return newRemoteExecutor(namespace, name), nil
}
