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

package configure

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type connectorExecutor struct {
	namespace  string
	name       string
	kubeConfig string
	keyFile    string
	user       string
	port       int
}

func newConnectorExecutor(opt Options) *connectorExecutor {
	return &connectorExecutor{
		namespace:  opt.Namespace,
		name:       opt.Name,
		kubeConfig: opt.KubeConfig,
		keyFile:    opt.KeyFile,
		user:       opt.User,
		port:       opt.Port,
	}
}

func (exe *connectorExecutor) GetName() string {
	return exe.name
}

func (exe *connectorExecutor) Execute() error {
	// Get config
	connector, err := config.GetConnector(exe.namespace, exe.name)
	if err != nil {
		return err
	}

	// Disallow editing vanilla fields for k8s Connector
	if connector.KubeConfig != "" && (exe.port != 0 || exe.keyFile != "" || exe.user != "") {
		return util.NewInputError("Connector " + exe.name + " is deployed on Kubernetes. You cannot add SSH details to this Connector")
	}

	// Disallow editing k8s fields for vanilla Connector
	if connector.SSH.Host != "" && exe.kubeConfig != "" {
		return util.NewInputError("Connector " + exe.name + " is not deployed on Kubernetes. You cannot add Kube Config details to this Connector")
	}

	if exe.keyFile != "" {
		connector.SSH.KeyFile = exe.keyFile
	}

	if exe.user != "" {
		connector.SSH.User = exe.user
	}

	if exe.port != 0 {
		connector.SSH.Port = exe.port
	}

	if exe.kubeConfig != "" {
		connector.KubeConfig = exe.kubeConfig
	}

	// Save config
	if err = config.UpdateConnector(exe.namespace, connector); err != nil {
		return err
	}

	return config.Flush()
}
