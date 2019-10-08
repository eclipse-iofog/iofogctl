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
	"strconv"

	"github.com/eclipse-iofog/iofog-go-sdk/pkg/client"
	types "github.com/eclipse-iofog/iofogctl/pkg/iofog/deploy"
)

// ApplicationData is data fetched from controller at init time
type ApplicationData struct {
	MicroserviceByName map[string]*client.MicroserviceInfo
	AgentsByName       map[string]*client.AgentInfo
	CatalogByID        map[int]*client.CatalogItemInfo
	RegistryByID       map[int]*client.RegistryInfo
	CatalogByName      map[string]*client.CatalogItemInfo
	FlowInfo           *client.FlowInfo
}

type remoteExecutor struct {
	controller         types.IofogController
	msvc               types.Microservice
	microserviceByName map[string]*client.MicroserviceInfo
	agentsByName       map[string]*client.AgentInfo
	catalogByID        map[int]*client.CatalogItemInfo
	catalogByName      map[string]*client.CatalogItemInfo
	registryByID       map[int]*client.RegistryInfo
	flowInfo           *client.FlowInfo
	client             *client.Client
	routes             []string
}

func newRemoteExecutor(controller types.IofogController, msvc types.Microservice) *remoteExecutor {
	exe := &remoteExecutor{
		controller: controller,
		msvc:       msvc,
	}

	return exe
}

// NewRemoteExecutorWithApplicationDataAndClient used by application deployment in order to reuse already initialised data
func NewRemoteExecutorWithApplicationDataAndClient(controller types.IofogController, msvc types.Microservice, appData ApplicationData, clt *client.Client) *remoteExecutor {
	exe := &remoteExecutor{
		controller:         controller,
		msvc:               msvc,
		client:             clt,
		microserviceByName: appData.MicroserviceByName,
		flowInfo:           appData.FlowInfo,
		catalogByID:        appData.CatalogByID,
		catalogByName:      appData.CatalogByName,
		agentsByName:       appData.AgentsByName,
		registryByID:       appData.RegistryByID,
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
	// Init remote resources
	if err = exe.init(); err != nil {
		return
	}

	// Validate microservice definition (routes, agents, etc.)
	if err = exe.validate(); err != nil {
		return
	}

	// Deploy microservice
	if _, err = exe.Deploy(); err != nil {
		return
	}
	return nil
}

func (exe *remoteExecutor) init() (err error) {
	exe.client, err = client.NewAndLogin(exe.controller.Endpoint, exe.controller.Email, exe.controller.Password)
	if err != nil {
		return
	}
	if exe.msvc.Flow == nil {
		return types.NewInputError("You must specify an application in order to deploy a microservice")
	}
	flowList, err := exe.client.GetAllFlows()
	if err != nil {
		return
	}
	for _, flow := range flowList.Flows {
		if flow.Name == *exe.msvc.Flow {
			exe.flowInfo = &flow
		}
	}
	if exe.flowInfo == nil {
		return types.NewInputError(fmt.Sprintf("Could not find application [%s]", *exe.msvc.Flow))
	}
	listMsvcs, err := exe.client.GetMicroservicesPerFlow(exe.flowInfo.ID)
	if err != nil {
		return
	}
	exe.microserviceByName = make(map[string]*client.MicroserviceInfo)
	for i := 0; i < len(listMsvcs.Microservices); i++ {
		exe.microserviceByName[listMsvcs.Microservices[i].Name] = &listMsvcs.Microservices[i]
		// If msvc already exists, set UUID
		if listMsvcs.Microservices[i].Name == exe.msvc.Name {
			if exe.msvc.UUID == "" {
				exe.msvc.UUID = listMsvcs.Microservices[i].UUID
			} else if exe.msvc.UUID != listMsvcs.Microservices[i].UUID {
				return types.NewConflictError(fmt.Sprintf("Cannot deploy microservice, there is a UUID mismatch. Controller UUID [%s], YAML UUID [%s]", listMsvcs.Microservices[i].UUID, exe.msvc.UUID))
			}
		}
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

func (exe *remoteExecutor) validate() error {
	// Validate routes
	routes, err := validateRoutes(exe.msvc.Routes, exe.microserviceByName)
	if err != nil {
		return err
	}

	exe.routes = routes

	// Validate microservice
	if err := ValidateMicroservice(exe.msvc, exe.agentsByName, exe.catalogByID, exe.registryByID); err != nil {
		return err
	}

	// Validate update
	if exe.msvc.UUID != "" {
		existingMsvc := exe.microserviceByName[exe.msvc.Name]
		if exe.msvc.Images.CatalogID != 0 && exe.msvc.Images.CatalogID != existingMsvc.CatalogItemID {
			return types.NewInputError(fmt.Sprintf("Cannot update a microservice catalog item"))
		}
		if exe.flowInfo != nil && exe.flowInfo.ID != existingMsvc.FlowID {
			return types.NewInputError(fmt.Sprintf("Cannot update a microservice application"))
		}
	}
	// TODO: Check if microservice already exists (Will fail on API call)
	return nil
}

func (exe *remoteExecutor) Deploy() (newMsvc *client.MicroserviceInfo, err error) {
	// Get catalog item
	catalogItem, err := SetUpCatalogItem(&exe.msvc, exe.catalogByID, exe.catalogByName, exe.client)
	if err != nil {
		return nil, err
	}
	var catalogItemID int
	if catalogItem != nil {
		catalogItemID = catalogItem.ID
	}

	// Get registry

	// Configure agent
	agent, err := ConfigureAgent(&exe.msvc, exe.agentsByName[exe.msvc.Agent.Name], exe.client)
	if err != nil {
		return nil, err
	}

	// Transform msvc config to JSON string
	config := ""
	if exe.msvc.Config != nil {
		byteconfig, err := json.Marshal(exe.msvc.Config)
		if err != nil {
			return nil, err
		}
		config = string(byteconfig)
	}

	var registryID int
	if exe.msvc.Images.Registry != "" {
		registryID, err = strconv.Atoi(exe.msvc.Images.Registry)
		if err != nil {
			registryID = client.RegistryTypeRegistryTypeIDDict[exe.msvc.Images.Registry]
		}
	}

	if exe.msvc.UUID != "" {
		// Update microservice
		return exe.update(config, agent.UUID, catalogItemID, registryID)
	}
	// Create microservice
	return exe.create(config, agent.UUID, catalogItemID, registryID)
}

func (exe *remoteExecutor) create(config, agentUUID string, catalogID, registryID int) (newMsvc *client.MicroserviceInfo, err error) {
	images := []client.CatalogImage{
		{ContainerImage: exe.msvc.Images.X86, AgentTypeID: client.AgentTypeAgentTypeIDDict["x86"]},
		{ContainerImage: exe.msvc.Images.ARM, AgentTypeID: client.AgentTypeAgentTypeIDDict["arm"]},
	}
	return exe.client.CreateMicroservice(client.MicroserviceCreateRequest{
		Config:         config,
		CatalogItemID:  catalogID,
		FlowID:         exe.flowInfo.ID,
		Name:           exe.msvc.Name,
		RootHostAccess: exe.msvc.RootHostAccess,
		Ports:          exe.msvc.Ports,
		Volumes:        exe.msvc.Volumes,
		RegistryID:     registryID,
		Env:            exe.msvc.Env,
		AgentUUID:      agentUUID,
		Routes:         exe.routes,
		Images:         images,
	})
}

func (exe *remoteExecutor) update(config, agentUUID string, catalogID, registryID int) (newMsvc *client.MicroserviceInfo, err error) {
	images := []client.CatalogImage{
		{ContainerImage: exe.msvc.Images.X86, AgentTypeID: client.AgentTypeAgentTypeIDDict["x86"]},
		{ContainerImage: exe.msvc.Images.ARM, AgentTypeID: client.AgentTypeAgentTypeIDDict["arm"]},
	}

	return exe.client.UpdateMicroservice(client.MicroserviceUpdateRequest{
		UUID:           exe.msvc.UUID,
		Config:         &config,
		Name:           &exe.msvc.Name,
		RootHostAccess: &exe.msvc.RootHostAccess,
		Ports:          exe.msvc.Ports,
		Volumes:        &exe.msvc.Volumes,
		Env:            exe.msvc.Env,
		AgentUUID:      &agentUUID,
		RegistryID:     &registryID,
		Routes:         exe.routes,
		Images:         images,
	})
}
