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

package deploycontroller

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
)

type kubernetesExecutor struct {
	opt *Options
}

func newKubernetesExecutor(opt *Options) *kubernetesExecutor {
	k := &kubernetesExecutor{}
	k.opt = opt
	return k
}

func (exe *kubernetesExecutor) Execute() (err error) {
	// Get Kubernetes cluster
	k8s, err := iofog.NewKubernetes(exe.opt.KubeConfig, exe.opt.Namespace)
	if err != nil {
		return
	}

	// Configure deploy
	if err = k8s.SetImages(exe.opt.Images); err != nil {
		return err
	}
	k8s.SetControllerIP(exe.opt.KubeControllerIP)

	// Update configuration before we try to deploy in case of failure
	configEntry, err := prepareUserAndSaveConfig(exe.opt)
	if err != nil {
		return
	}

	// Create controller on cluster
	endpoint, err := k8s.CreateController(iofog.User{
		Name:     configEntry.IofogUser.Name,
		Surname:  configEntry.IofogUser.Surname,
		Email:    configEntry.IofogUser.Email,
		Password: configEntry.IofogUser.Password,
	})
	if err != nil {
		return
	}

	// Update configuration
	configEntry.Endpoint = endpoint
	if err = config.UpdateController(exe.opt.Namespace, configEntry); err != nil {
		return
	}

	return config.Flush()
}
