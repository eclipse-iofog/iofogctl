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

package deploymicroservice

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
	microservice apps.Microservice
	controller   apps.IofogController
}

func (exe remoteExecutor) GetName() string {
	return exe.microservice.Name
}

func (exe remoteExecutor) Execute() error {
	util.SpinStart(fmt.Sprintf("Deploying microservice %s", exe.GetName()))
	return apps.DeployMicroservice(exe.controller, exe.microservice)
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
	var microservice apps.Microservice
	if err = yaml.UnmarshalStrict(opt.Yaml, &microservice); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}

	// Check Name is lowercase
	if err := util.IsLowerAlphanumeric(opt.Name); err != nil {
		return nil, err
	}

	if len(opt.Name) > 0 {
		microservice.Name = opt.Name
	}

	endpoint, err := ns.ControlPlane.GetControllerEndpoint()
	if err != nil {
		return
	}

	return remoteExecutor{
		controller: apps.IofogController{
			Endpoint: endpoint,
			Email:    ns.ControlPlane.IofogUser.Email,
			Password: ns.ControlPlane.IofogUser.Password,
		},
		microservice: microservice}, nil
}
