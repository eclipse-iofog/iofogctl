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
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	jsoniter "github.com/json-iterator/go"
)

func validatePortMapping(port *MicroservicePortMapping, agentsByName map[string]*client.AgentInfo) error {
	isPublic := port.Public != 0
	if !isPublic && port.Host != "" {
		return NewInputError("Cannot specify a port host without specifying a public port number")
	}
	if port.Protocol != "" {
		port.Protocol = strings.ToLower(port.Protocol)
		protocol := port.Protocol
		if protocol != "tcp" && protocol != "http" {
			return NewInputError(fmt.Sprintf("Protocol %s is not supported. Valid protocols are tcp and http\n", protocol))
		}
	}
	if port.Host != "" {
		if port.Host != client.DefaultRouterName {
			agent, found := agentsByName[port.Host]
			if !found {
				return NewNotFoundError(fmt.Sprintf("Could not find port host %s\n", port.Host))
			}
			port.Host = agent.UUID
		}
	}
	return nil
}

func validateMicroservice(
	msvc *Microservice,
	agentsByName map[string]*client.AgentInfo,
	catalogByID map[int]*client.CatalogItemInfo,
	registryByID map[int]*client.RegistryInfo) (err error) {
	// Validate ports and update host
	for idx := range msvc.Container.Ports {
		if err := validatePortMapping(&msvc.Container.Ports[idx], agentsByName); err != nil {
			return err
		}
	}

	// Validate microservice
	var agent *client.AgentInfo
	var catalogItem *client.CatalogItemInfo
	var foundAgent, foundCatalogItem bool
	if agent, foundAgent = agentsByName[msvc.Agent.Name]; !foundAgent {
		return NewNotFoundError(fmt.Sprintf("Could not find agent: %s", msvc.Agent.Name))
	}
	if catalogItem, foundCatalogItem = catalogByID[msvc.Images.CatalogID]; msvc.Images.CatalogID > 0 && !foundCatalogItem {
		return NewNotFoundError(fmt.Sprintf("Could not find catalog item: %d", msvc.Images.CatalogID))
	}
	registryID, _ := strconv.Atoi(msvc.Images.Registry)
	if _, foundRegistry := registryByID[registryID]; msvc.Images.Registry != "" && !foundRegistry {
		if _, foundRegistry := client.RegistryTypeRegistryTypeIDDict[msvc.Images.Registry]; msvc.Images.Registry != "" && !foundRegistry {
			return NewInputError(fmt.Sprintf("Invalid registry: %s", msvc.Images.Registry))
		}
	}

	// Check if msvc image for the agent type is provided
	if msvc.Images.CatalogID > 0 {
		found := false
		for _, img := range catalogItem.Images {
			if img.AgentTypeID == agent.FogType {
				found = true
				break
			}
		}
		if !found {
			return NewInputError(fmt.Sprintf("Microservice %s does not have a valid image for the Agent %s", msvc.Name, agent.Name))
		}
	} else {
		switch agent.FogType {
		case 1:
			if msvc.Images.X86 == "" {
				return NewInputError(fmt.Sprintf("Microservice %s does not have a valid image for the Agent %s", msvc.Name, agent.Name))
			}
		case 2:
			if msvc.Images.ARM == "" {
				return NewInputError(fmt.Sprintf("Microservice %s does not have a valid image for the Agent %s", msvc.Name, agent.Name))
			}
		}
	}

	// TODO: Check if microservice name already exists in another flow (Will fail on API call)
	return nil
}

func validateRoutes(routes []string, microserviceByName map[string]*client.MicroserviceInfo) (routesUUIDs []string, err error) { // nolint:deadcode,unused
	// Validate routes
	for _, route := range routes {
		msvc, foundTo := microserviceByName[route]
		if !foundTo {
			return routesUUIDs, NewNotFoundError(fmt.Sprintf("Could not find microservice [%s] required by a route", route))
		}
		routesUUIDs = append(routesUUIDs, msvc.UUID)
	}
	return routesUUIDs, nil
}

func configureAgent(msvc *Microservice, agent *client.AgentInfo, clt *client.Client) (*client.AgentInfo, error) {
	return clt.UpdateAgent(&client.AgentUpdateRequest{
		UUID: agent.UUID,
		AgentConfiguration: client.AgentConfiguration{
			DockerURL:                 msvc.Agent.Config.DockerURL,
			DiskLimit:                 msvc.Agent.Config.DiskLimit,
			DiskDirectory:             msvc.Agent.Config.DiskDirectory,
			MemoryLimit:               msvc.Agent.Config.MemoryLimit,
			CPULimit:                  msvc.Agent.Config.CPULimit,
			LogLimit:                  msvc.Agent.Config.LogLimit,
			LogDirectory:              msvc.Agent.Config.LogDirectory,
			LogFileCount:              msvc.Agent.Config.LogFileCount,
			StatusFrequency:           msvc.Agent.Config.StatusFrequency,
			ChangeFrequency:           msvc.Agent.Config.ChangeFrequency,
			DeviceScanFrequency:       msvc.Agent.Config.DeviceScanFrequency,
			BluetoothEnabled:          msvc.Agent.Config.BluetoothEnabled,
			WatchdogEnabled:           msvc.Agent.Config.WatchdogEnabled,
			AbstractedHardwareEnabled: msvc.Agent.Config.AbstractedHardwareEnabled,
		},
	})
}

func setUpCatalogItem(
	msvc *Microservice,
	catalogByID map[int]*client.CatalogItemInfo,
	catalogByName map[string]*client.CatalogItemInfo,
	clt *client.Client) (catalogItem *client.CatalogItemInfo, err error) {
	// No catalog item
	if msvc.Images.CatalogID == 0 {
		return
	}
	// Catalog item, and no image
	if msvc.Images.CatalogID > 0 && msvc.Images.X86 == "" && msvc.Images.ARM == "" {
		return catalogByID[msvc.Images.CatalogID], nil
	}
	catalogImages := []client.CatalogImage{
		{ContainerImage: msvc.Images.X86, AgentTypeID: client.AgentTypeAgentTypeIDDict["x86"]},
		{ContainerImage: msvc.Images.ARM, AgentTypeID: client.AgentTypeAgentTypeIDDict["arm"]},
	}
	registryID, ok := client.RegistryTypeRegistryTypeIDDict[msvc.Images.Registry]
	if !ok {
		registryID = 1 // Remote by default
	}
	// Get possible exisiting catalog item
	catalogItemName := fmt.Sprintf("%s_catalog", msvc.Name)
	var found bool
	if msvc.Images.CatalogID == 0 {
		catalogItem, found = catalogByName[catalogItemName]
	} else {
		catalogItem, found = catalogByID[msvc.Images.CatalogID]
	}
	// Update catalog item if needed
	if found {
		// Check if catalog item needs to be updated
		if catalogItemNeedsUpdate(catalogItem, catalogImages, registryID) {
			if msvc.Images.CatalogID != 0 {
				return nil, NewInputError("Cannot update a microservice catalog item")
			}
			// Delete catalog item
			if err := clt.DeleteCatalogItem(catalogItem.ID); err != nil {
				return nil, err
			}
			// Create new catalog item
			catalogItem, err = clt.CreateCatalogItem(&client.CatalogItemCreateRequest{
				Name:        catalogItemName,
				Description: fmt.Sprintf("Catalog item for msvc %s", msvc.Name),
				Images:      catalogImages,
				RegistryID:  registryID,
			})
			if err != nil {
				return nil, err
			}
		}
	} else if msvc.Images.CatalogID == 0 { // If not found and no catalog item id, create a new one
		// Create new catalog item
		catalogItem, err = clt.CreateCatalogItem(&client.CatalogItemCreateRequest{
			Name:        catalogItemName,
			Description: fmt.Sprintf("Catalog item for msvc %s", msvc.Name),
			Images:      catalogImages,
			RegistryID:  registryID,
		})
		if err != nil {
			return nil, err
		}
	} else { // Not found, and catalog item id specified
		return nil, NewNotFoundError(fmt.Sprintf("Could not find specified catalog item, ID: %d", msvc.Images.CatalogID))
	}
	return catalogItem, err
}

func createRoutes(routes []Route, microserviceByName map[string]*client.MicroserviceInfo, clt *client.Client) (err error) { // nolint:deadcode,unused
	for _, route := range routes {
		fromMsvc := microserviceByName[route.From]
		toMsvc := microserviceByName[route.To]
		if err = clt.CreateMicroserviceRoute(fromMsvc.UUID, toMsvc.UUID); err != nil {
			return
		}
	}
	return nil
}

func catalogItemNeedsUpdate(catalogItem *client.CatalogItemInfo, catalogImages []client.CatalogImage, registry int) bool {
	if catalogItem.RegistryID != registry || len(catalogImages) != len(catalogItem.Images) {
		return true
	}

	currentImagesPerAgentType := make(map[int]string)
	for _, currentImage := range catalogItem.Images {
		currentImagesPerAgentType[currentImage.AgentTypeID] = currentImage.ContainerImage
	}

	for _, image := range catalogImages {
		if currentImage, found := currentImagesPerAgentType[image.AgentTypeID]; !found || currentImage != image.ContainerImage {
			return true
		}
	}

	return false
}

func mapMicroserviceToClientMicroserviceRequest(microservice *Microservice) (request client.MicroserviceCreateRequest, err error) {
	// Transform msvc config to JSON string
	config := ""
	if microservice.Config != nil {
		byteconfig, err := jsoniter.Marshal(microservice.Config)
		if err != nil {
			return client.MicroserviceCreateRequest{}, err
		}
		config = string(byteconfig)
	}

	var registryID int
	if microservice.Images.Registry != "" {
		registryID, err = strconv.Atoi(microservice.Images.Registry)
		if err != nil {
			registryID = client.RegistryTypeRegistryTypeIDDict[microservice.Images.Registry]
		}
	}

	images := []client.CatalogImage{
		{ContainerImage: microservice.Images.X86, AgentTypeID: client.AgentTypeAgentTypeIDDict["x86"]},
		{ContainerImage: microservice.Images.ARM, AgentTypeID: client.AgentTypeAgentTypeIDDict["arm"]},
	}
	volumes := mapVolumes(microservice.Container.Volumes)
	if volumes == nil {
		volumes = &[]client.MicroserviceVolumeMapping{}
	}
	envs := mapEnvs(microservice.Container.Env)
	if envs == nil {
		envs = &[]client.MicroserviceEnvironment{}
	}
	extraHosts := mapExtraHosts(microservice.Container.ExtraHosts)
	if extraHosts == nil {
		extraHosts = &[]client.MicroserviceExtraHost{}
	}
	return client.MicroserviceCreateRequest{
		Config:         config,
		CatalogItemID:  microservice.Images.CatalogID,
		Name:           microservice.Name,
		RootHostAccess: microservice.Container.RootHostAccess,
		Ports:          mapPorts(microservice.Container.Ports),
		Volumes:        *volumes,
		Env:            *envs,
		ExtraHosts:     *extraHosts,
		RegistryID:     registryID,
		AgentName:      microservice.Agent.Name,
		Commands:       microservice.Container.Commands,
		Images:         images,
	}, nil
}

func mapRouteToClientRouteRequest(route Route) client.ApplicationRouteCreateRequest {
	return client.ApplicationRouteCreateRequest{
		From: route.From,
		To:   route.To,
		Name: route.Name,
	}
}

func mapMicroservicesToClientMicroserviceRequests(microservices []Microservice) (result []client.MicroserviceCreateRequest, err error) {
	result = make([]client.MicroserviceCreateRequest, 0)
	for idx := range microservices {
		msvc := &microservices[idx]
		request, err := mapMicroserviceToClientMicroserviceRequest(msvc)
		if err != nil {
			return result, err
		}
		result = append(result, request)
	}
	return
}

func mapRoutesToClientRouteRequests(routes []Route) (result []client.ApplicationRouteCreateRequest) {
	result = make([]client.ApplicationRouteCreateRequest, 0)
	if len(routes) == 0 {
		return
	}
	for _, route := range routes {
		result = append(result, mapRouteToClientRouteRequest(route))
	}
	return
}

func mapVariablesToClientVariables(variables []TemplateVariable) (result []client.TemplateVariable) {
	if len(variables) == 0 {
		return
	}
	for _, variable := range variables {
		clientVariable := client.TemplateVariable{
			Key:          variable.Key,
			Description:  variable.Description,
			DefaultValue: variable.DefaultValue,
			Value:        variable.Value,
		}
		result = append(result, clientVariable)
	}
	return
}

func mapTemplateToClientTemplate(template *ApplicationTemplate) (result *client.ApplicationTemplate) {
	if template != nil {
		result = &client.ApplicationTemplate{
			Name: template.Name,
		}
		result.Variables = mapVariablesToClientVariables(template.Variables)
	}
	return
}
