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
	"encoding/json"
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	deploymicroservice "github.com/eclipse-iofog/iofogctl/internal/deploy/microservice"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type remoteExecutor struct {
	namespace          string
	app                config.Application
	microserviceByName map[string]*client.MicroserviceInfo
	client             *client.Client
	agentsByName       map[string]*client.AgentInfo
	catalogByID        map[int]*client.CatalogItemInfo
	catalogByName      map[string]*client.CatalogItemInfo
}

func microserviceArrayToMap(a []config.Microservice) (result map[string]*client.MicroserviceInfo) {
	result = make(map[string]*client.MicroserviceInfo)
	for i := 0; i < len(a); i++ {
		// No need to fill information, we only need to know if the name exists
		result[a[i].Name] = &client.MicroserviceInfo{}
	}
	return
}

func newRemoteExecutor(namespace string, app config.Application) *remoteExecutor {
	exe := &remoteExecutor{
		namespace:          namespace,
		app:                app,
		microserviceByName: microserviceArrayToMap(app.Microservices),
	}

	return exe
}

func (exe *remoteExecutor) GetName() string {
	return exe.app.Name
}

//
// Deploy application using remote controller
//
func (exe *remoteExecutor) Execute() (err error) {
	// Get Control Plane
	controlPlane, err := config.GetControlPlane(exe.namespace)
	if err != nil || len(controlPlane.Controllers) == 0 {
		util.PrintError("You must deploy a Controller to a namespace before deploying any Agents")
		return
	}

	// Init remote resources
	if err = exe.init(&controlPlane.Controllers[0], controlPlane.IofogUser); err != nil {
		return
	}

	// Validate application definition (routes, agents, etc.)
	if err = exe.validate(); err != nil {
		return
	}

	// Deploy application
	if err = exe.deploy(); err != nil {
		return
	}
	return nil
}

func (exe *remoteExecutor) init(controller *config.Controller, user config.IofogUser) (err error) {
	exe.client = client.New(controller.Endpoint)
	if err = exe.client.Login(client.LoginRequest{Email: user.Email, Password: user.Password}); err != nil {
		return
	}
	listAgents, err := exe.client.ListAgents()
	if err != nil {
		return
	}
	exe.agentsByName = make(map[string]*client.AgentInfo)
	for i := 0; i < len(listAgents.Agents); i++ {
		exe.agentsByName[listAgents.Agents[i].Name] = &listAgents.Agents[i]
	}

	listCatalog, err := exe.client.GetCatalog()
	if err != nil {
		return
	}
	exe.catalogByID = make(map[int]*client.CatalogItemInfo)
	exe.catalogByName = make(map[string]*client.CatalogItemInfo)
	for i := 0; i < len(listCatalog.CatalogItems); i++ {
		exe.catalogByID[listCatalog.CatalogItems[i].ID] = &listCatalog.CatalogItems[i]
		exe.catalogByName[listCatalog.CatalogItems[i].Name] = &listCatalog.CatalogItems[i]
	}
	return
}

func (exe *remoteExecutor) validate() (err error) {
	// Validate routes
	for _, route := range exe.app.Routes {
		if _, foundFrom := exe.microserviceByName[route.From]; !foundFrom {
			return util.NewNotFoundError(fmt.Sprintf("Could not find origin microservice for the route %v", route))
		}
		if _, foundTo := exe.microserviceByName[route.To]; !foundTo {
			return util.NewNotFoundError(fmt.Sprintf("Could not find destination microservice for the route %v", route))
		}
	}

	// Validate microservice
	for _, msvc := range exe.app.Microservices {
		if err = deploymicroservice.ValidateMicroservice(msvc, exe.agentsByName, exe.catalogByID); err != nil {
			return
		}
	}

	// TODO: Check if application alredy exists
	return nil
}

func (exe *remoteExecutor) createRoutes() (err error) {
	for _, route := range exe.app.Routes {
		fromMsvc, _ := exe.microserviceByName[route.From]
		toMsvc, _ := exe.microserviceByName[route.To]
		if err = exe.client.CreateMicroserviceRoute(fromMsvc.UUID, toMsvc.UUID); err != nil {
			return
		}
	}
	return nil
}

func (exe *remoteExecutor) deploy() (err error) {
	// Create flow
	flow, err := exe.client.CreateFlow(exe.app.Name, fmt.Sprintf("Flow for application: %s", exe.app.Name))
	if err != nil {
		return
	}

	// Create microservices
	for _, msvc := range exe.app.Microservices {
		util.SpinStart(fmt.Sprintf("Deploying microservice %s", msvc.Name))

		// Configure agent
		agent, err := deploymicroservice.ConfigureAgent(&msvc, exe.agentsByName[msvc.Agent.Name], exe.client)
		if err != nil {
			return err
		}

		// Get catalog item
		catalogItem, err := deploymicroservice.SetUpCatalogItem(&msvc, exe.catalogByID, exe.catalogByName, exe.client)
		if err != nil {
			return err
		}

		// Transform msvc config to JSON string
		config := ""
		if msvc.Config != nil {
			byteconfig, err := json.Marshal(msvc.Config)
			if err != nil {
				return err
			}
			config = string(byteconfig)
		}

		// Create microservice
		msvcInfo, err := exe.client.CreateMicroservice(client.MicroserviceCreateRequest{
			Config:         config,
			CatalogItemID:  catalogItem.ID,
			FlowID:         flow.ID,
			Name:           msvc.Name,
			RootHostAccess: msvc.RootHostAccess,
			Ports:          msvc.Ports,
			Volumes:        msvc.Volumes,
			Env:            msvc.Env,
			AgentUUID:      agent.UUID,
		})
		if err != nil {
			return err
		}
		// Update msvc map with UUID for routing
		exe.microserviceByName[msvc.Name].UUID = msvcInfo.UUID
	}

	// Create Routes
	if err = exe.createRoutes(); err != nil {
		return
	}

	// Start flow
	util.SpinStart("Starting flow")
	active := true
	if flow, err = exe.client.UpdateFlow(&client.FlowUpdateRequest{
		IsActivated: &active,
		ID:          flow.ID,
	}); err != nil {
		return
	}
	return nil
}
