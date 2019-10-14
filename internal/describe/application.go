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
	deploy "github.com/eclipse-iofog/iofog-go-sdk/pkg/apps"
	"github.com/eclipse-iofog/iofog-go-sdk/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
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

func (exe *applicationExecutor) init(controlPlane config.ControlPlane) (err error) {
	// TODO: Replace controller[0] with variable in controlPlane
	exe.client = client.New(controlPlane.Controllers[0].Endpoint)
	if err = exe.client.Login(client.LoginRequest{Email: controlPlane.IofogUser.Email, Password: controlPlane.IofogUser.Password}); err != nil {
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

	// Filter system microservices
	for _, msvc := range msvcListResponse.Microservices {
		if util.IsSystemMsvc(msvc) {
			continue
		}
		exe.msvcs = append(exe.msvcs, msvc)
	}
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
	// Get Control Plane config details
	controlPlane, err := config.GetControlPlane(exe.namespace)
	if err != nil {
		return err
	}
	// Check Controller exists
	if len(controlPlane.Controllers) == 0 {
		return util.NewInputError("This namespace does not have a Controller. You must first deploy a Controller describing Applications.")
	}
	// Fetch data
	if err = exe.init(controlPlane); err != nil {
		return err
	}

	yamlMsvcs := []deploy.Microservice{}
	yamlRoutes := []deploy.Route{}

	for _, msvc := range exe.msvcs {
		yamlMsvc, err := MapClientMicroserviceToDeployMicroservice(&msvc, exe.client)
		if err != nil {
			return err
		}
		for _, route := range msvc.Routes {
			yamlRoutes = append(yamlRoutes, deploy.Route{
				From: yamlMsvc.Name,
				To:   exe.msvcPerID[route].Name,
			})
		}
		// Remove fields
		yamlMsvc.Routes = nil
		yamlMsvc.Flow = nil
		yamlMsvcs = append(yamlMsvcs, *yamlMsvc)
	}

	application := deploy.Application{
		Name:          exe.flow.Name,
		Microservices: yamlMsvcs,
		Routes:        yamlRoutes,
		ID:            exe.flow.ID,
	}

	header := deploy.Header{
		Kind: deploy.ApplicationKind,
		Metadata: deploy.HeaderMetadata{
			Namespace: exe.namespace,
		},
		Spec: application,
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
