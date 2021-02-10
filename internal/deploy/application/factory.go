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
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/v2/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
	"gopkg.in/yaml.v2"
)

type Options struct {
	Namespace string
	Yaml      []byte
	Name      string
}

type remoteExecutor struct {
	namespace   string
	application *rsc.Application
}

func (exe *remoteExecutor) GetName() string {
	return exe.application.Name
}

func (exe *remoteExecutor) Execute() error {
	util.SpinStart(fmt.Sprintf("Deploying Application %s", exe.GetName()))

	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	controlPlane, err := ns.GetControlPlane()
	if err != nil {
		return err
	}

	// Check Controller exists
	if len(controlPlane.GetControllers()) == 0 {
		return util.NewInputError("This namespace does not have a Controller. You must first deploy a Controller before deploying Applications")
	}

	endpoint, err := controlPlane.GetEndpoint()
	if err != nil {
		return err
	}

	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	controller := apps.IofogController{
		Endpoint: endpoint,
		Email:    controlPlane.GetUser().Email,
		Password: controlPlane.GetUser().Password,
		Token:    clt.GetAccessToken(),
	}
	return apps.DeployApplication(controller, exe.application)
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	// Check the namespace exists
	if _, err = config.GetNamespace(opt.Namespace); err != nil {
		return exe, err
	}
	// Unmarshal file
	application := rsc.Application{}
	if err = yaml.UnmarshalStrict(opt.Yaml, &application); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}
	for _, route := range application.Routes {
		if err := util.IsLowerAlphanumeric("Route", route.Name); err != nil {
			return nil, err
		}
	}

	if len(opt.Name) > 0 {
		application.Name = opt.Name
	}

	if err := util.IsLowerAlphanumeric("Application", application.Name); err != nil {
		return nil, err
	}

	// Update default msvc values
	for idx := range application.Microservices {
		if application.Microservices[idx].Images == nil {
			return nil, util.NewInputError(fmt.Sprintf("The microservice %s does not have an image. You must provide a valid microservice image to deploy",
				application.Microservices[idx].Name))
		}
		if application.Microservices[idx].Images.Registry == "" {
			application.Microservices[idx].Images.Registry = "remote"
		}
	}

	return &remoteExecutor{
		namespace:   opt.Namespace,
		application: &application}, nil
}
