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

package describe

import (
	"github.com/eclipse-iofog/iofog-go-sdk/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	deploy "github.com/eclipse-iofog/iofogctl/pkg/iofog/deploy"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type microserviceExecutor struct {
	namespace string
	name      string
	filename  string
	client    *client.Client
	msvc      *client.MicroserviceInfo
}

func newMicroserviceExecutor(namespace, name, filename string) *microserviceExecutor {
	a := &microserviceExecutor{}
	a.namespace = namespace
	a.name = name
	a.filename = filename
	return a
}

func (exe *microserviceExecutor) init(controlPlane config.ControlPlane) (err error) {
	// TODO: Replace controller[0] with variable in controlPlane
	exe.client = client.New(controlPlane.Controllers[0].Endpoint)
	if err = exe.client.Login(client.LoginRequest{Email: controlPlane.IofogUser.Email, Password: controlPlane.IofogUser.Password}); err != nil {
		return
	}
	exe.msvc, err = exe.client.GetMicroserviceByName(exe.name)
	return
}

func (exe *microserviceExecutor) GetName() string {
	return exe.name
}

func (exe *microserviceExecutor) Execute() error {
	// Get Control Plane config details
	controlPlane, err := config.GetControlPlane(exe.namespace)
	if err != nil {
		return err
	}
	// Check Controller exists
	if len(controlPlane.Controllers) == 0 {
		return util.NewInputError("This namespace does not have a Controller. You must first deploy a Controller describing Microservices.")
	}
	// Fetch data
	if err = exe.init(controlPlane); err != nil {
		return err
	}

	if util.IsSystemMsvc(*(exe.msvc)) {
		return nil
	}

	yamlMsvc, err := MapClientMicroserviceToDeployMicroservice(exe.msvc, exe.client)
	if err != nil {
		return err
	}

	header := deploy.Header{
		Kind: deploy.MicroserviceKind,
		Metadata: deploy.HeaderMetadata{
			Namespace: exe.namespace,
		},
		Spec: yamlMsvc,
	}

	if exe.filename == "" {
		if err = util.Print(header); err != nil {
			return err
		}
	} else {
		if err = util.FPrint(header, exe.filename); err != nil {
			return err
		}
	}
	return nil
}
