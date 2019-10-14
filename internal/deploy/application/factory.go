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

package deployapplication

import (
	"fmt"

	apps "github.com/eclipse-iofog/iofog-go-sdk/pkg/apps"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"gopkg.in/yaml.v2"
)

type Options struct {
	Namespace string
	Yaml      []byte
	Name      string
}

type remoteExecutor struct {
	application apps.Application
	controller  apps.IofogController
}

func (exe remoteExecutor) GetName() string {
	return exe.application.Name
}

func (exe remoteExecutor) Execute() error {
	util.SpinStart(fmt.Sprintf("Deploying application %s", exe.GetName()))
	return apps.DeployApplication(exe.controller, exe.application)
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	// Check the namespace exists
	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return exe, err
	}

	// Check Controller exists
	if len(ns.ControlPlane.Controllers) == 0 {
		return exe, util.NewInputError("This namespace does not have a Controller. You must first deploy a Controller before deploying Applications")
	}

	// Unmarshal file
	application := apps.Application{}
	if err = yaml.Unmarshal(opt.Yaml, &application); err != nil {
		err = util.NewInputError("Could not unmarshall\n" + err.Error())
		return
	}

	if len(opt.Name) > 0 {
		application.Name = opt.Name
	}

	return remoteExecutor{
		controller: apps.IofogController{
			Endpoint: ns.ControlPlane.Controllers[0].Endpoint,
			Email:    ns.ControlPlane.IofogUser.Email,
			Password: ns.ControlPlane.IofogUser.Password,
		},
		application: application}, nil
}
