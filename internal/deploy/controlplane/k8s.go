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

package deploycontrolplane

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/client"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
)

type kubernetesExecutor struct {
	namespace string
	ctrl      config.Controller
}

func newKubernetesExecutor(namespace string, ctrl config.Controller) *kubernetesExecutor {
	k := &kubernetesExecutor{}
	k.namespace = namespace
	k.ctrl = ctrl
	return k
}

func (exe *kubernetesExecutor) GetName() string {
	return exe.ctrl.Name
}

func (exe *kubernetesExecutor) Execute() (err error) {
	// Get Kubernetes cluster
	k8s, err := install.NewKubernetes(exe.ctrl.KubeConfig, exe.namespace)
	if err != nil {
		return
	}

	// Configure deploy
	if err = k8s.SetImages(exe.ctrl.Images); err != nil {
		return err
	}
	k8s.SetControllerIP(exe.ctrl.KubeControllerIP)

	// Update configuration before we try to deploy in case of failure
	configEntry, err := prepareUserAndSaveConfig(exe.namespace, exe.ctrl)
	if err != nil {
		return
	}

	// Create controller on cluster
	endpoint, err := k8s.CreateController(client.User{
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
	if err = config.UpdateController(exe.namespace, configEntry); err != nil {
		return
	}

	return config.Flush()
}
