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
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/client"
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

func (exe *microserviceExecutor) init(controller *config.Controller) (err error) {
	exe.client = client.New(controller.Endpoint)
	if err = exe.client.Login(client.LoginRequest{Email: controller.IofogUser.Email, Password: controller.IofogUser.Password}); err != nil {
		return
	}
	exe.msvc, err = exe.client.GetMicroserviceByName(exe.name)
	return
}

func (exe *microserviceExecutor) GetName() string {
	return exe.name
}

func (exe *microserviceExecutor) Execute() error {
	// Get controller config details
	controllers, err := config.GetControllers(exe.namespace)
	if err != nil {
		return err
	}
	// Check Controller exists
	if len(controllers) == 0 {
		return util.NewInputError("This namespace does not have a Controller. You must first deploy a Controller describing Microservices.")
	}
	// Fetch data
	if err = exe.init(&controllers[0]); err != nil {
		return err
	}

	yamlMsvc, err := MapClientMicroserviceToConfigMicroservice(exe.msvc, exe.client)
	if err != nil {
		return err
	}

	if exe.filename == "" {
		if err = util.Print(yamlMsvc); err != nil {
			return err
		}
	} else {
		if err = util.FPrint(yamlMsvc, exe.filename); err != nil {
			return err
		}
	}
	return nil
}
