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

package connect

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
)

type kubernetesExecutor struct {
	opt *Options
}

func newKubernetesExecutor(opt *Options) *kubernetesExecutor {
	k := &kubernetesExecutor{}
	k.opt = opt
	return k
}

func (exe *kubernetesExecutor) GetName() string {
	return exe.opt.Name
}

func (exe *kubernetesExecutor) Execute() (err error) {
	// Instantiate Kubernetes cluster object
	k8s, err := install.NewKubernetes(exe.opt.KubeFile, exe.opt.Namespace)
	if err != nil {
		return err
	}

	// Check the resources exist in K8s namespace
	err = k8s.ExistsInNamespace(exe.opt.Namespace)
	if err != nil {
		return err
	}

	// Get Controller endpoint
	endpoint, err := k8s.GetControllerEndpoint()
	if err != nil {
		return err
	}

	// Establish connection
	err = connect(exe.opt, endpoint)
	if err != nil {
		return err
	}
	return config.Flush()
}
