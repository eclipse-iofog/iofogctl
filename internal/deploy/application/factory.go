/*
 *  *******************************************************************************
 *  * Copyright (c) 2020 Edgeworx, Inc.
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

	apps "github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/apps"
	"github.com/eclipse-iofog/iofogctl/v2/internal"
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/internal/execute"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
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
	controlPlane, err := ns.GetControlPlane()
	if err != nil {
		return exe, err
	}

	// Check Controller exists
	if len(controlPlane.GetControllers()) == 0 {
		return exe, util.NewInputError("This namespace does not have a Controller. You must first deploy a Controller before deploying Applications")
	}

	// Unmarshal file
	application := apps.Application{}
	if err = yaml.UnmarshalStrict(opt.Yaml, &application); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}

	if len(opt.Name) > 0 {
		application.Name = opt.Name
	}

	// Update default msvc values
	for idx := range application.Microservices {
		if application.Microservices[idx].Images.Registry == "" {
			application.Microservices[idx].Images.Registry = "remote"
		}
	}

	endpoint, err := controlPlane.GetEndpoint()
	if err != nil {
		return
	}

	clt, err := internal.NewControllerClient(opt.Namespace)
	if err != nil {
		return
	}

	return remoteExecutor{
		controller: apps.IofogController{
			Endpoint: endpoint,
			Email:    controlPlane.GetUser().Email,
			Password: controlPlane.GetUser().Password,
			Token:    clt.GetAccessToken(),
		},
		application: application}, nil
}
