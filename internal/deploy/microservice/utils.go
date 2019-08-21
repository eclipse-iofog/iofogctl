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
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func MicroserviceArrayToMap(a []config.Microservice) (result map[string]*config.Microservice) {
	result = make(map[string]*config.Microservice)
	for i := 0; i < len(a); i++ {
		result[a[i].Name] = &a[i]
	}
	return
}

func ValidateMicroservice(msvc config.Microservice, agentsByName map[string]*client.AgentInfo, catalogByID map[int]*client.CatalogItemInfo) (err error) {
	// Validate microservice
	if _, foundAgent := agentsByName[msvc.Agent.Name]; !foundAgent {
		return util.NewNotFoundError(fmt.Sprintf("Could not find agent: %s", msvc.Agent.Name))
	}
	if _, foundCatalogItem := catalogByID[msvc.Images.CatalogID]; msvc.Images.CatalogID > 0 && !foundCatalogItem {
		return util.NewNotFoundError(fmt.Sprintf("Could not find catalog item: %d", msvc.Images.CatalogID))
	}

	// TODO: Check if microservice already exists (Will fail on API call)
	return nil
}

func validateRoutes(routes []string, microserviceByName map[string]*client.MicroserviceInfo) (routesUUIDs []string, err error) {
	// Validate routes
	for _, route := range routes {
		msvc, foundTo := microserviceByName[route]
		if !foundTo {
			return routesUUIDs, util.NewNotFoundError(fmt.Sprintf("Could not find microservice [%s] required by a route", route))
		}
		routesUUIDs = append(routesUUIDs, msvc.UUID)
	}
	return routesUUIDs, nil
}

func ConfigureAgent(msvc *config.Microservice, agent *client.AgentInfo, clt *client.Client) (*client.AgentInfo, error) {
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

func SetUpCatalogItem(msvc *config.Microservice, catalogByID map[int]*client.CatalogItemInfo, catalogByName map[string]*client.CatalogItemInfo, clt *client.Client) (catalogItem *client.CatalogItemInfo, err error) {
	catalogImages := []client.CatalogImage{
		{ContainerImage: msvc.Images.X86, AgentTypeID: client.AgentTypeAgentTypeIDDict["x86"]},
		{ContainerImage: msvc.Images.ARM, AgentTypeID: client.AgentTypeAgentTypeIDDict["arm"]},
	}
	registryID, ok := client.RegistryTypeRegistryTypeIDDict[msvc.Images.Registry]
	if !ok {
		registryID = 1 // Remote by default
	}
	if msvc.Images.CatalogID == 0 {
		catalogItemName := fmt.Sprintf("%s_catalog", msvc.Name)
		var found bool
		catalogItem, found = catalogByName[catalogItemName]
		if found == true {
			// Check if catalog item needs to be updated
			if catalogItemNeedsUpdate(catalogItem, catalogImages, registryID) {
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
		} else {
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
	} else {
		catalogItem = catalogByID[msvc.Images.CatalogID]
	}
	return
}

func CreateRoutes(routes []config.Route, microserviceByName map[string]*client.MicroserviceInfo, clt *client.Client) (err error) {
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
