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

package describe

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	iutil "github.com/eclipse-iofog/iofogctl/v2/internal/util"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type applicationExecutor struct {
	namespace string
	name      string
	filename  string
	flow      *client.FlowInfo
	client    *client.Client
	msvcs     []client.MicroserviceInfo
	routes    []client.Route
	msvcPerID map[string]*client.MicroserviceInfo
}

func newApplicationExecutor(namespace, name, filename string) *applicationExecutor {
	a := &applicationExecutor{}
	a.namespace = namespace
	a.name = name
	a.filename = filename
	return a
}

func (exe *applicationExecutor) init() (err error) {
	exe.client, err = iutil.NewControllerClient(exe.namespace)
	if err != nil {
		return
	}

	routeList, err := exe.client.ListRoutes()
	if err != nil {
		return err
	}
	exe.routes = routeList.Routes

	application, err := exe.client.GetApplicationByName(exe.name)
	// If notfound error, try legacy
	if _, ok := err.(*client.NotFoundError); err != nil && ok {
		return exe.initLegacy()
	}
	// TODO: Use Application instead of flow
	exe.flow = &client.FlowInfo{
		Name:        application.Name,
		IsActivated: application.IsActivated,
		Description: application.Description,
		IsSystem:    application.IsSystem,
		UserID:      application.UserID,
		ID:          application.ID,
	}
	msvcListResponse, err := exe.client.GetMicroservicesByApplication(exe.name)
	if err != nil {
		return err
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
	// Fetch data
	if err := exe.init(); err != nil {
		return err
	}

	yamlMsvcs := []rsc.Microservice{}
	yamlRoutes := []rsc.Route{}

	for _, msvc := range exe.msvcs {
		yamlMsvc, err := MapClientMicroserviceToDeployMicroservice(&msvc, exe.client)
		if err != nil {
			return err
		}
		// Remove fields
		yamlMsvc.Flow = nil
		yamlMsvcs = append(yamlMsvcs, *yamlMsvc)
	}

	for _, route := range exe.routes {
		from, okSrc := exe.msvcPerID[route.SourceMicroserviceUUID]
		to, okDest := exe.msvcPerID[route.DestMicroserviceUUID]
		if okSrc {
			if !okDest {
				return util.NewNotFoundError(fmt.Sprintf("Route %s contains a destination microservice that could not be found in the application", route.Name))
			}
			yamlRoutes = append(yamlRoutes, rsc.Route{
				Name: route.Name,
				From: from.Name,
				To:   to.Name,
			})
		}
	}

	application := rsc.Application{
		Name:          exe.flow.Name,
		Microservices: yamlMsvcs,
		Routes:        yamlRoutes,
		ID:            exe.flow.ID,
	}

	header := config.Header{
		APIVersion: config.LatestAPIVersion,
		Kind:       config.ApplicationKind,
		Metadata: config.HeaderMetadata{
			Namespace: exe.namespace,
			Name:      exe.name,
		},
		Spec: application,
	}

	if exe.filename == "" {
		if err := util.Print(header); err != nil {
			return err
		}
	} else {
		if err := util.FPrint(header, exe.filename); err != nil {
			return err
		}
	}
	return nil
}
