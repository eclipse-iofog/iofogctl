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

	"github.com/eclipse-iofog/iofog-go-sdk/pkg/client"
)

func validateMicroservice(msvc Microservice, agentsByName map[string]*client.AgentInfo, catalogByID map[int]*client.CatalogItemInfo, registryByID map[int]*client.RegistryInfo) (err error) {
	// Validate microservice
	if _, foundAgent := agentsByName[msvc.Agent.Name]; !foundAgent {
		return NewNotFoundError(fmt.Sprintf("Could not find agent: %s", msvc.Agent.Name))
	}
	if _, foundCatalogItem := catalogByID[msvc.Images.CatalogID]; msvc.Images.CatalogID > 0 && !foundCatalogItem {
		return NewNotFoundError(fmt.Sprintf("Could not find catalog item: %d", msvc.Images.CatalogID))
	}
	registryID, _ := strconv.Atoi(msvc.Images.Registry)
	if _, foundRegistry := registryByID[registryID]; msvc.Images.Registry != "" && !foundRegistry {
		if _, foundRegistry := client.RegistryTypeRegistryTypeIDDict[msvc.Images.Registry]; msvc.Images.Registry != "" && !foundRegistry {
			return NewInputError(fmt.Sprintf("Invalid registry: %s", msvc.Images.Registry))
		}
	}

	// TODO: Check if microservice name already exists in another flow (Will fail on API call)
	return nil
}

func validateRoutes(routes []string, microserviceByName map[string]*client.MicroserviceInfo) (routesUUIDs []string, err error) {
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

func setUpCatalogItem(msvc *Microservice, catalogByID map[int]*client.CatalogItemInfo, catalogByName map[string]*client.CatalogItemInfo, clt *client.Client) (catalogItem *client.CatalogItemInfo, err error) {
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
	if found == true {
		// Check if catalog item needs to be updated
		if catalogItemNeedsUpdate(catalogItem, catalogImages, registryID) {
			if msvc.Images.CatalogID != 0 {
				return nil, NewInputError(fmt.Sprintf("Cannot update a microservice catalog item"))
			}
			// Delete catalog item
			if err = clt.DeleteCatalogItem(catalogItem.ID); err != nil {
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
	return
}

func createRoutes(routes []Route, microserviceByName map[string]*client.MicroserviceInfo, clt *client.Client) (err error) {
	for _, route := range routes {
		fromMsvc, _ := microserviceByName[route.From]
		toMsvc, _ := microserviceByName[route.To]
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
