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
	"strconv"

	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	jsoniter "github.com/json-iterator/go"
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

type microserviceExecutor struct {
	controller         IofogController
	msvc               Microservice
	microserviceByName map[string]*client.MicroserviceInfo
	agentsByName       map[string]*client.AgentInfo
	catalogByID        map[int]*client.CatalogItemInfo
	catalogByName      map[string]*client.CatalogItemInfo
	registryByID       map[int]*client.RegistryInfo
	flowInfo           *client.FlowInfo
	client             *client.Client
	routes             []string
}

func newMicroserviceExecutor(controller IofogController, msvc Microservice) *microserviceExecutor {
	exe := &microserviceExecutor{
		controller: controller,
		msvc:       msvc,
	}

	return exe
}

// newMicroserviceExecutorWithApplicationDataAndClient used by application deployment in order to reuse already initialised data
func newMicroserviceExecutorWithApplicationDataAndClient(controller IofogController, msvc Microservice, appData ApplicationData, clt *client.Client) *microserviceExecutor {
	exe := &microserviceExecutor{
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

func (exe *microserviceExecutor) execute() (err error) {
	// Init remote resources
	if err = exe.init(); err != nil {
		return
	}

	// Validate microservice definition (routes, agents, etc.)
	if err = exe.validate(); err != nil {
		return
	}

	// Deploy microservice
	if _, err = exe.deploy(); err != nil {
		return
	}
	return nil
}

func (exe *microserviceExecutor) init() (err error) {
	if exe.controller.Token != "" {
		exe.client, err = client.NewWithToken(client.Options{Endpoint: exe.controller.Endpoint}, exe.controller.Token)
	} else {
		exe.client, err = client.NewAndLogin(client.Options{Endpoint: exe.controller.Endpoint}, exe.controller.Email, exe.controller.Password)
	}
	if err != nil {
		return
	}
	if exe.msvc.Flow == nil {
		return NewInputError("You must specify an application in order to deploy a microservice")
	}
	exe.flowInfo, err = exe.client.GetFlowByName(*exe.msvc.Flow)
	if err != nil {
		return err
	}
	if exe.flowInfo == nil {
		return NewInputError(fmt.Sprintf("Could not find application [%s]", *exe.msvc.Flow))
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
				return NewConflictError(fmt.Sprintf("Cannot deploy microservice, there is a UUID mismatch. Controller UUID [%s], YAML UUID [%s]", listMsvcs.Microservices[i].UUID, exe.msvc.UUID))
			}
		}
	}
	listAgents, err := exe.client.ListAgents(client.ListAgentsRequest{})
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

func (exe *microserviceExecutor) validate() error {
	// Validate routes
	routes, err := validateRoutes(exe.msvc.Routes, exe.microserviceByName)
	if err != nil {
		return err
	}

	exe.routes = routes

	// Validate microservice
	if err := validateMicroservice(&exe.msvc, exe.agentsByName, exe.catalogByID, exe.registryByID); err != nil {
		return err
	}

	// Validate update
	if exe.msvc.UUID != "" {
		existingMsvc := exe.microserviceByName[exe.msvc.Name]
		if exe.msvc.Images.CatalogID != 0 && exe.msvc.Images.CatalogID != existingMsvc.CatalogItemID {
			return NewInputError(fmt.Sprintf("Cannot update a microservice catalog item"))
		}
		if exe.flowInfo != nil && exe.flowInfo.ID != existingMsvc.FlowID {
			return NewInputError(fmt.Sprintf("Cannot update a microservice application"))
		}
	}
	// TODO: Check if microservice already exists (Will fail on API call)
	return nil
}

func (exe *microserviceExecutor) deploy() (newMsvc *client.MicroserviceInfo, err error) {
	// Get catalog item
	catalogItem, err := setUpCatalogItem(&exe.msvc, exe.catalogByID, exe.catalogByName, exe.client)
	if err != nil {
		return nil, err
	}
	var catalogItemID int
	if catalogItem != nil {
		catalogItemID = catalogItem.ID
	}

	// Get registry

	// Configure agent
	agent, err := configureAgent(&exe.msvc, exe.agentsByName[exe.msvc.Agent.Name], exe.client)
	if err != nil {
		return nil, err
	}

	// Transform msvc config to JSON string
	config := ""
	if exe.msvc.Config != nil {
		byteconfig, err := jsoniter.Marshal(exe.msvc.Config)
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

func (exe *microserviceExecutor) create(config, agentUUID string, catalogID, registryID int) (newMsvc *client.MicroserviceInfo, err error) {
	images := []client.CatalogImage{
		{ContainerImage: exe.msvc.Images.X86, AgentTypeID: client.AgentTypeAgentTypeIDDict["x86"]},
		{ContainerImage: exe.msvc.Images.ARM, AgentTypeID: client.AgentTypeAgentTypeIDDict["arm"]},
	}
	volumes := mapVolumes(exe.msvc.Container.Volumes)
	if volumes == nil {
		volumes = &[]client.MicroserviceVolumeMapping{}
	}
	envs := mapEnvs(exe.msvc.Container.Env)
	if envs == nil {
		envs = &[]client.MicroserviceEnvironment{}
	}
	extraHosts := mapExtraHosts(exe.msvc.Container.ExtraHosts)
	if extraHosts == nil {
		extraHosts = &[]client.MicroserviceExtraHost{}
	}
	return exe.client.CreateMicroservice(client.MicroserviceCreateRequest{
		Config:         config,
		CatalogItemID:  catalogID,
		FlowID:         exe.flowInfo.ID,
		Name:           exe.msvc.Name,
		RootHostAccess: exe.msvc.Container.RootHostAccess,
		Ports:          mapPorts(exe.msvc.Container.Ports),
		Volumes:        *volumes,
		Env:            *envs,
		ExtraHosts:     *extraHosts,
		RegistryID:     registryID,
		AgentUUID:      agentUUID,
		Routes:         exe.routes,
		Commands:       exe.msvc.Container.Commands,
		Images:         images,
	})
}

func (exe *microserviceExecutor) update(config, agentUUID string, catalogID, registryID int) (newMsvc *client.MicroserviceInfo, err error) {
	images := []client.CatalogImage{
		{ContainerImage: exe.msvc.Images.X86, AgentTypeID: client.AgentTypeAgentTypeIDDict["x86"]},
		{ContainerImage: exe.msvc.Images.ARM, AgentTypeID: client.AgentTypeAgentTypeIDDict["arm"]},
	}

	var cmdPointer *[]string
	if exe.msvc.Container.Commands != nil {
		cmdPointer = &exe.msvc.Container.Commands
	}

	return exe.client.UpdateMicroservice(client.MicroserviceUpdateRequest{
		UUID:           exe.msvc.UUID,
		Config:         &config,
		CatalogItemID:  catalogID,
		Name:           &exe.msvc.Name,
		RootHostAccess: &exe.msvc.Container.RootHostAccess,
		Ports:          mapPorts(exe.msvc.Container.Ports),
		Volumes:        mapVolumes(exe.msvc.Container.Volumes),
		Env:            mapEnvs(exe.msvc.Container.Env),
		ExtraHosts:     mapExtraHosts(exe.msvc.Container.ExtraHosts),
		AgentUUID:      &agentUUID,
		RegistryID:     &registryID,
		Routes:         exe.routes,
		Commands:       cmdPointer,
		Images:         images,
		Rebuild:        exe.msvc.Rebuild,
	})
}

func mapPorts(in []MicroservicePortMapping) (out []client.MicroservicePortMapping) {
	for _, port := range in {
		out = append(out, client.MicroservicePortMapping(port))
	}
	return
}

func mapVolumes(in *[]MicroserviceVolumeMapping) *[]client.MicroserviceVolumeMapping {
	if in == nil {
		return nil
	}

	out := make([]client.MicroserviceVolumeMapping, 0)
	for _, vol := range *in {
		out = append(out, client.MicroserviceVolumeMapping(vol))
	}
	return &out
}

func mapEnvs(in *[]MicroserviceEnvironment) *[]client.MicroserviceEnvironment {
	if in == nil {
		return nil
	}

	out := make([]client.MicroserviceEnvironment, 0)
	for _, env := range *in {
		out = append(out, client.MicroserviceEnvironment(env))
	}
	return &out
}

func mapExtraHosts(in *[]MicroserviceExtraHost) *[]client.MicroserviceExtraHost {
	if in == nil {
		return nil
	}

	out := make([]client.MicroserviceExtraHost, 0)
	for _, eH := range *in {
		out = append(out, client.MicroserviceExtraHost(eH))
	}
	return &out
}
