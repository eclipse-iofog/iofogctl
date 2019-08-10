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
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type applicationExecutor struct {
	namespace string
	name      string
	filename  string
	flow      *client.FlowInfo
	client    *client.Client
	msvcs     []client.MicroserviceInfo
	msvcPerID map[string]*client.MicroserviceInfo
}

func newApplicationExecutor(namespace, name, filename string) *applicationExecutor {
	a := &applicationExecutor{}
	a.namespace = namespace
	a.name = name
	a.filename = filename
	return a
}

func (exe *applicationExecutor) init(controller *config.Controller) (err error) {
	exe.client = client.New(controller.Endpoint)
	if err = exe.client.Login(client.LoginRequest{Email: controller.IofogUser.Email, Password: controller.IofogUser.Password}); err != nil {
		return
	}
	exe.flow, err = exe.client.GetFlowByName(exe.name)
	if err != nil {
		return
	}
	msvcListResponse, err := exe.client.GetMicroservicesPerFlow(exe.flow.ID)
	if err != nil {
		return
	}
	exe.msvcs = msvcListResponse.Microservices
	exe.msvcPerID = make(map[string]*client.MicroserviceInfo)
	for i := 0; i < len(exe.msvcs); i++ {
		exe.msvcPerID[exe.msvcs[i].UUID] = &exe.msvcs[i]
	}
	return
}

func (exe *applicationExecutor) GetName() string {
	return exe.name
}

func (exe *applicationExecutor) Execute() error {
	// Get controller config details
	controllers, err := config.GetControllers(exe.namespace)
	if err != nil {
		return err
	}
	if len(controllers) != 1 {
		errMessage := fmt.Sprintf("This namespace contains %d Controller(s), you must have one, and only one.", len(controllers))
		return util.NewInputError(errMessage)
	}
	// Fetch data
	if err = exe.init(&controllers[0]); err != nil {
		return err
	}

	yamlMsvcs := []config.Microservice{}
	yamlRoutes := []config.Route{}

	for _, msvc := range exe.msvcs {
		yamlMsvc, err := MapClientMicroserviceToConfigMicroservice(&msvc, exe.client)
		if err != nil {
			return err
		}
		for _, route := range msvc.Routes {
			yamlRoutes = append(yamlRoutes, config.Route{
				From: yamlMsvc.Name,
				To:   exe.msvcPerID[route].Name,
			})
		}
		// Remove fields
		yamlMsvc.Routes = nil
		yamlMsvc.Flow = nil
		yamlMsvcs = append(yamlMsvcs, *yamlMsvc)
	}

	application := config.Application{
		Name:          exe.flow.Name,
		Microservices: yamlMsvcs,
		Routes:        yamlRoutes,
	}

	if exe.filename == "" {
		if err = util.Print(application); err != nil {
			return err
		}
	} else {
		if err = util.FPrint(application, exe.filename); err != nil {
			return err
		}
	}
	return nil
}
