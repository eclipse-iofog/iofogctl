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
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
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
	k8s, err := iofog.NewKubernetes(exe.opt.KubeConfig)
	if err != nil {
		return
	}

	// Configure deploy
	k8s.SetImages(exe.opt.Images)
	k8s.SetControllerIP(exe.opt.KubeControllerIP)

	var configUser config.IofogUser
	// Check existing user
	ctrl, err := config.GetController(exe.opt.Namespace, exe.opt.Name)
	if err == nil {
		// Use existing user
		configUser = ctrl.IofogUser
	} else {
		// Generate new user
		configUser = config.NewUser()
	}
	// Assign user
	user := iofog.User{
		Name:     configUser.Name,
		Surname:  configUser.Surname,
		Email:    configUser.Email,
		Password: configUser.Password,
	}

	// Update configuration before we try to deploy in case of failure
	configEntry := config.Controller{
		Name:       exe.opt.Name,
		KubeConfig: exe.opt.KubeConfig,
		IofogUser: config.IofogUser{
			Name:     user.Name,
			Surname:  user.Surname,
			Email:    user.Email,
			Password: user.Password,
		},
		Created: util.NowUTC(),
	}
	if err = config.UpdateController(exe.opt.Namespace, configEntry); err != nil {
		return
	}
	if err = config.Flush(); err != nil {
		return err
	}

	// Create controller on cluster
	endpoint, err := k8s.CreateController(user)
	if err != nil {
		return
	}

	// Update configuration
	configEntry.Endpoint = endpoint
	if err = config.UpdateController(exe.opt.Namespace, configEntry); err != nil {
		return
	}

	fmt.Printf("\nController %s/%s successfully deployed.\n", exe.opt.Namespace, exe.opt.Name)

	return config.Flush()
}
