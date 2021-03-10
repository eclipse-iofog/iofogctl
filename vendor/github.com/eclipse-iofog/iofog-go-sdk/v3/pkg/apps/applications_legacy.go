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

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
)

func (exe *applicationExecutor) initLegacy() (err error) {
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
	return err
}

func (exe *applicationExecutor) validateLegacy() (err error) {
	// Validate routes
	for _, route := range exe.app.Routes {
		if _, foundFrom := exe.microserviceByName[route.From]; !foundFrom {
			return NewNotFoundError(fmt.Sprintf("Could not find origin microservice for the route %v", route))
		}
		if _, foundTo := exe.microserviceByName[route.To]; !foundTo {
			return NewNotFoundError(fmt.Sprintf("Could not find destination microservice for the route %v", route))
		}
		if route.Name == "" {
			route.Name = route.From + "-to-" + route.To
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
		if err := exe.client.UpdateRoute(client.Route{
			Name:                   route.Name,
			SourceMicroserviceUUID: microserviceByName[route.From].UUID,
			DestMicroserviceUUID:   microserviceByName[route.To].UUID,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (exe *applicationExecutor) updateLegacy() (err error) {
	description := fmt.Sprintf("Flow for application: %s", exe.app.Name)
	// Update and stop flow
	active := false
	if _, err := exe.client.UpdateFlow(&client.FlowUpdateRequest{
		Name:        &exe.app.Name,
		Description: &description,
		IsActivated: &active,
		ID:          exe.flowInfo.ID,
	}); err != nil {
		return err
	}

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
	for idx := range listMsvcs.Microservices {
		msvc := &listMsvcs.Microservices[idx]
		catalogItem, foundCatalogItem := exe.catalogByID[msvc.CatalogItemID]
		// If !foundCatalogItem -> Catalog item not returned in init -> We cannot edit it.
		isSystem := msvc.CatalogItemID != 0 && (!foundCatalogItem || catalogItem.Category == "SYSTEM")
		if _, found := yamlMicroservicesPerName[msvc.Name]; !found && !isSystem {
			if err := exe.client.DeleteMicroservice(msvc.UUID); err != nil {
				return err
			}
		}
	}

	// Deploy microservices
	for _, msvc := range yamlMicroservicesPerName {
		// Force deletion of all routes
		msvcExecutor := newMicroserviceExecutorWithApplicationDataAndClient(
			exe.controller,
			msvc,
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

func (exe *applicationExecutor) createLegacy() (err error) {
	description := fmt.Sprintf("Flow for application: %s", exe.app.Name)
	// Create flow
	flow, err := exe.client.CreateFlow(exe.app.Name, description)
	if err != nil {
		return err
	}

	exe.flowInfo = flow

	// Create microservices
	for idx := range exe.app.Microservices {
		msvcExecutor := newMicroserviceExecutorWithApplicationDataAndClient(
			exe.controller,
			&exe.app.Microservices[idx],
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

func (exe *applicationExecutor) deployLegacy() (err error) {
	// Validate application definition (routes, agents, etc.)
	if err = exe.initLegacy(); err != nil {
		return
	}
	// Validate application definition (routes, agents, etc.)
	if err = exe.validateLegacy(); err != nil {
		return
	}

	if exe.flowInfo == nil {
		if err := exe.createLegacy(); err != nil {
			return err
		}
	} else {
		if err := exe.updateLegacy(); err != nil {
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
