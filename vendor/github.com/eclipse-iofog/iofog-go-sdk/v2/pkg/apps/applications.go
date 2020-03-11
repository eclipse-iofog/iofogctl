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

package apps

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
)

type iofogUser struct {
	email    string
	password string
}

type applicationExecutor struct {
	controller         IofogController
	app                Application
	microserviceByName map[string]*client.MicroserviceInfo
	client             *client.Client
	flowInfo           *client.FlowInfo
	agentsByName       map[string]*client.AgentInfo
	catalogByID        map[int]*client.CatalogItemInfo
	catalogByName      map[string]*client.CatalogItemInfo
	registryByID       map[int]*client.RegistryInfo
}

func microserviceArrayToClientMap(a []Microservice) (result map[string]*client.MicroserviceInfo) {
	result = make(map[string]*client.MicroserviceInfo)
	for i := 0; i < len(a); i++ {
		// No need to fill information, we only need to know if the name exists
		result[a[i].Name] = &client.MicroserviceInfo{}
	}
	return
}

func newApplicationExecutor(controller IofogController, app Application) *applicationExecutor {
	exe := &applicationExecutor{
		controller:         controller,
		app:                app,
		microserviceByName: microserviceArrayToClientMap(app.Microservices),
	}

	return exe
}

func (exe *applicationExecutor) execute() (err error) {
	// Init remote resources
	if err = exe.init(); err != nil {
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

func (exe *applicationExecutor) init() (err error) {
	if exe.controller.Token != "" {
		exe.client, err = client.NewWithToken(client.Options{Endpoint: exe.controller.Endpoint}, exe.controller.Token)
	} else {
		exe.client, err = client.NewAndLogin(client.Options{Endpoint: exe.controller.Endpoint}, exe.controller.Email, exe.controller.Password)
	}
	if err != nil {
		return
	}

	// Look for exisiting flow
	var flowInfo *client.FlowInfo
	if exe.app.ID != 0 {
		flowInfo, err = exe.client.GetFlowByID(exe.app.ID)
	} else {
		flowInfo, err = exe.client.GetFlowByName(exe.app.Name)
	}

	// If not notfound error, return error
	if _, ok := err.(*client.NotFoundError); err != nil && !ok {
		return err
	}

	exe.flowInfo = flowInfo

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
	listRegistries, err := exe.client.ListRegistries()
	if err != nil {
		return
	}
	exe.registryByID = make(map[int]*client.RegistryInfo)
	for i := 0; i < len(listRegistries.Registries); i++ {
		exe.registryByID[listRegistries.Registries[i].ID] = &listRegistries.Registries[i]
	}
	return
}

func (exe *applicationExecutor) validate() (err error) {
	// Validate routes
	for _, route := range exe.app.Routes {
		if _, foundFrom := exe.microserviceByName[route.From]; !foundFrom {
			return NewNotFoundError(fmt.Sprintf("Could not find origin microservice for the route %v", route))
		}
		if _, foundTo := exe.microserviceByName[route.To]; !foundTo {
			return NewNotFoundError(fmt.Sprintf("Could not find destination microservice for the route %v", route))
		}
	}

	// Validate microservice
	for idx := range exe.app.Microservices {
		if err = validateMicroservice(&exe.app.Microservices[idx], exe.agentsByName, exe.catalogByID, exe.registryByID); err != nil {
			return
		}
	}

	// TODO: Check if application alredy exists
	return nil
}

func (exe *applicationExecutor) createRoutes(microserviceByName map[string]*client.MicroserviceInfo) (err error) {
	for _, route := range exe.app.Routes {
		fromMsvc, _ := microserviceByName[route.From]
		toMsvc, _ := microserviceByName[route.To]
		if err = exe.client.CreateMicroserviceRoute(fromMsvc.UUID, toMsvc.UUID); err != nil {
			return err
		}
	}
	return nil
}

func (exe *applicationExecutor) update() (err error) {
	description := fmt.Sprintf("Flow for application: %s", exe.app.Name)
	// Update and stop flow
	active := false
	exe.client.UpdateFlow(&client.FlowUpdateRequest{
		Name:        &exe.app.Name,
		Description: &description,
		IsActivated: &active,
	})

	existingMicroservicesPerName := make(map[string]*client.MicroserviceInfo)
	listMsvcs, err := exe.client.GetMicroservicesPerFlow(exe.flowInfo.ID)
	if err != nil {
		return err
	}
	for idx := range listMsvcs.Microservices {
		existingMicroservicesPerName[listMsvcs.Microservices[idx].Name] = &listMsvcs.Microservices[idx]
	}

	yamlMicroservicesPerName := make(map[string]*Microservice)
	for idx := range exe.app.Microservices {
		// Set flow
		exe.app.Microservices[idx].Flow = &exe.app.Name
		// Set possible UUID
		if msvc, found := existingMicroservicesPerName[exe.app.Microservices[idx].Name]; found {
			exe.app.Microservices[idx].UUID = msvc.UUID
		}
		yamlMicroservicesPerName[exe.app.Microservices[idx].Name] = &exe.app.Microservices[idx]
	}

	// Delete all uneeded microservices
	for _, msvc := range listMsvcs.Microservices {
		catalogItem, foundCatalogItem := exe.catalogByID[msvc.CatalogItemID]
		// If !foundCatalogItem -> Catalog item not returned in init -> We cannot edit it.
		isSystem := msvc.CatalogItemID != 0 && (!foundCatalogItem || catalogItem.Category == "SYSTEM")
		if _, found := yamlMicroservicesPerName[msvc.Name]; !found && !isSystem {
			if err = exe.client.DeleteMicroservice(msvc.UUID); err != nil {
				return err
			}
		}
	}

	// Deploy microservices
	for _, msvc := range yamlMicroservicesPerName {
		// Force deletion of all routes
		msvc.Routes = []string{}
		msvcExecutor := newMicroserviceExecutorWithApplicationDataAndClient(
			exe.controller,
			*msvc,
			ApplicationData{
				MicroserviceByName: existingMicroservicesPerName,
				AgentsByName:       exe.agentsByName,
				CatalogByID:        exe.catalogByID,
				CatalogByName:      exe.catalogByName,
				RegistryByID:       exe.registryByID,
				FlowInfo:           exe.flowInfo,
			},
			exe.client,
		)
		newMsvc, err := msvcExecutor.deploy()
		if err != nil {
			return err
		}
		existingMicroservicesPerName[newMsvc.Name] = newMsvc
	}

	// create routes
	return exe.createRoutes(existingMicroservicesPerName)
}

func (exe *applicationExecutor) create() (err error) {
	description := fmt.Sprintf("Flow for application: %s", exe.app.Name)
	// Create flow
	flow, err := exe.client.CreateFlow(exe.app.Name, description)
	if err != nil {
		return err
	}

	exe.flowInfo = flow

	// Create microservices
	for _, msvc := range exe.app.Microservices {
		msvcExecutor := newMicroserviceExecutorWithApplicationDataAndClient(
			exe.controller,
			msvc,
			ApplicationData{
				MicroserviceByName: exe.microserviceByName,
				AgentsByName:       exe.agentsByName,
				CatalogByID:        exe.catalogByID,
				CatalogByName:      exe.catalogByName,
				FlowInfo:           exe.flowInfo,
			},
			exe.client,
		)
		newMsvc, err := msvcExecutor.deploy()
		if err != nil {
			return err
		}

		exe.microserviceByName[newMsvc.Name] = newMsvc
	}

	// Create Routes
	return exe.createRoutes(exe.microserviceByName)
}

func (exe *applicationExecutor) deploy() (err error) {
	if exe.flowInfo == nil {
		if err = exe.create(); err != nil {
			return err
		}
	} else {
		if err = exe.update(); err != nil {
			return err
		}
	}

	// Start flow
	active := true
	if _, err = exe.client.UpdateFlow(&client.FlowUpdateRequest{
		IsActivated: &active,
		ID:          exe.flowInfo.ID,
	}); err != nil {
		return err
	}
	return nil
}
