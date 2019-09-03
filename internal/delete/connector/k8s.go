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
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
)

type kubernetesExecutor struct {
	namespace string
	name      string
}

func newKubernetesExecutor(namespace, name string) *kubernetesExecutor {
	exe := &kubernetesExecutor{}
	exe.namespace = namespace
	exe.name = name
	return exe
}

func (exe *kubernetesExecutor) GetName() string {
	return exe.name
}

func (exe *kubernetesExecutor) Execute() error {
	// Find the requested controller
	cnct, err := config.GetConnector(exe.namespace, exe.name)
	if err != nil {
		return err
	}

	// Instantiate Kubernetes object
	k8s, err := install.NewKubernetes(cnct.KubeConfig, exe.namespace)

	// Delete Connector on cluster
	err = k8s.DeleteConnector(exe.name)
	if err != nil {
		return err
	}

	// Clear Connector from Controller
	if err = deleteConnectorFromController(exe.namespace, cnct.Host); err != nil {
		return err
	}

	return nil
}
