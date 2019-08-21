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
	"encoding/json"
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type remoteExecutor struct {
	namespace          string
	msvc               config.Microservice
	microserviceByName map[string]*client.MicroserviceInfo
	client             *client.Client
	agentsByName       map[string]*client.AgentInfo
	catalogByID        map[int]*client.CatalogItemInfo
	catalogByName      map[string]*client.CatalogItemInfo
	routes             []string
}

func newRemoteExecutor(namespace string, msvc config.Microservice) *remoteExecutor {
	exe := &remoteExecutor{
		namespace: namespace,
		msvc:      msvc,
	}

	return exe
}

func (exe *remoteExecutor) GetName() string {
	return exe.msvc.Name
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

	if exe.msvc.Flow == nil {
		return util.NewInputError("You must specify a flow in order to deploy a microservice")
	}
	listMsvcs, err := exe.client.GetMicroservicesPerFlow(*exe.msvc.Flow)
	if err != nil {
		return
	}
	exe.microserviceByName = make(map[string]*client.MicroserviceInfo)
	for i := 0; i < len(listMsvcs.Microservices); i++ {
		exe.microserviceByName[listMsvcs.Microservices[i].Name] = &listMsvcs.Microservices[i]
	}
	return
}

func (exe *remoteExecutor) validate() error {
	// Validate routes
	routes, err := validateRoutes(exe.msvc.Routes, exe.microserviceByName)
	if err != nil {
		return err
	}

	exe.routes = routes

	// Validate microservice
	if err := ValidateMicroservice(exe.msvc, exe.agentsByName, exe.catalogByID); err != nil {
		return err
	}

	// TODO: Check if microservice already exists (Will fail on API call)
	return nil
}

func (exe *remoteExecutor) deploy() (err error) {
	// Create microservice
	util.SpinStart(fmt.Sprintf("Deploying microservice %s", exe.msvc.Name))

	// Configure agent
	agent, err := ConfigureAgent(&exe.msvc, exe.agentsByName[exe.msvc.Agent.Name], exe.client)
	if err != nil {
		return err
	}

	// Get catalog item
	catalogItem, err := SetUpCatalogItem(&exe.msvc, exe.catalogByID, exe.catalogByName, exe.client)
	if err != nil {
		return err
	}

	// Transform msvc config to JSON string
	config := ""
	if exe.msvc.Config != nil {
		byteconfig, err := json.Marshal(exe.msvc.Config)
		if err != nil {
			return err
		}
		config = string(byteconfig)
	}

	// Create microservice
	_, err = exe.client.CreateMicroservice(client.MicroserviceCreateRequest{
		Config:         config,
		CatalogItemID:  catalogItem.ID,
		FlowID:         *exe.msvc.Flow,
		Name:           exe.msvc.Name,
		RootHostAccess: exe.msvc.RootHostAccess,
		Ports:          exe.msvc.Ports,
		Volumes:        exe.msvc.Volumes,
		Env:            exe.msvc.Env,
		AgentUUID:      agent.UUID,
		Routes:         exe.routes,
	})
	return err
}
