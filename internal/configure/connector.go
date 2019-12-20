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
	namespace   string
	name        string
	kubeConfig  string
	host        string
	keyFile     string
	user        string
	port        int
	useDetached bool
}

func newConnectorExecutor(opt Options) *connectorExecutor {
	return &connectorExecutor{
		namespace:   opt.Namespace,
		name:        opt.Name,
		kubeConfig:  opt.KubeConfig,
		host:        opt.Host,
		keyFile:     opt.KeyFile,
		user:        opt.User,
		port:        opt.Port,
		useDetached: opt.UseDetached,
	}
}

func (exe *connectorExecutor) GetName() string {
	return exe.name
}

func (exe *connectorExecutor) Execute() error {
	var connector config.Connector
	var err error
	if exe.useDetached {
		connector, err = config.GetDetachedConnector(exe.name)
	} else {
		connector, err = config.GetConnector(exe.namespace, exe.name)
	}
	if err != nil {
		return err
	}

	// Disallow editing host if already exists
	if connector.Host != "" && exe.host != "" {
		util.NewInputError("Cannot edit existing host address of Connector. Can only add host address where it doesn't exist after running connect command")
	}

	// Disallow editing vanilla fields for k8s Connector
	if connector.Kube.Config != "" && (exe.port != 0 || exe.keyFile != "" || exe.user != "") {
		return util.NewInputError("Connector " + exe.name + " is deployed on Kubernetes. You cannot add SSH details to this Connector")
	}

	// Disallow editing k8s fields for vanilla Connector
	if connector.Host != "" && exe.kubeConfig != "" {
		return util.NewInputError("Connector " + exe.name + " has a host address already which suggests it is not on Kubernetes. You cannot add Kube Config details to this Connector")
	}

	// Only add/overwrite values provided
	if exe.host != "" {
		connector.Host = exe.host
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
		connector.Kube.Config = exe.kubeConfig
	}

	// Add port if not specified or existing
	if connector.Host != "" && connector.SSH.Port == 0 {
		connector.SSH.Port = 22
	}

	// Save config
	if exe.useDetached {
		return config.UpdateDetachedConnector(connector)
	}
	if err = config.UpdateConnector(exe.namespace, connector); err != nil {
		return err
	}

	return config.Flush()
}
