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

type controllerExecutor struct {
	namespace  string
	name       string
	kubeConfig string
	keyFile    string
	user       string
	port       int
}

func newControllerExecutor(opt Options) *controllerExecutor {
	return &controllerExecutor{
		namespace:  opt.Namespace,
		name:       opt.Name,
		kubeConfig: opt.KubeConfig,
		keyFile:    opt.KeyFile,
		user:       opt.User,
		port:       opt.Port,
	}
}

func (exe *controllerExecutor) GetName() string {
	return exe.name
}

func (exe *controllerExecutor) Execute() error {
	// Get config
	controller, err := config.GetController(exe.namespace, exe.name)
	if err != nil {
		return err
	}

	// Disallow editing vanilla fields for k8s Controller
	if controller.Kube.Config != "" && (exe.port != 0 || exe.keyFile != "" || exe.user != "") {
		return util.NewInputError("Controller " + exe.name + " is deployed on Kubernetes. You cannot add SSH details to this Controller")
	}

	// Disallow editing k8s fields for vanilla Controller
	if controller.Host != "" && exe.kubeConfig != "" {
		return util.NewInputError("Controller " + exe.name + " is not deployed on Kubernetes. You cannot add Kube Config details to this Controller")
	}

	if exe.keyFile != "" {
		controller.SSH.KeyFile = exe.keyFile
	}

	if exe.user != "" {
		controller.SSH.User = exe.user
	}

	if exe.port != 0 {
		controller.SSH.Port = exe.port
	}

	if exe.kubeConfig != "" {
		controller.Kube.Config = exe.kubeConfig
	}

	// Save config
	if err = config.UpdateController(exe.namespace, controller); err != nil {
		return err
	}

	return config.Flush()
}
