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
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type remoteExecutor struct {
	namespace          string
	opt                *config.Application
	microserviceByName map[string]*config.Microservice
	client             *client.Client
	agentsByName       map[string]*client.AgentInfo
	catalogByID        map[int]*client.CatalogItemInfo
	catalogByName      map[string]*client.CatalogItemInfo
}

func microserviceArrayToMap(a []config.Microservice) (result map[string]*config.Microservice) {
	result = make(map[string]*config.Microservice)
	for i := 0; i < len(a); i++ {
		result[a[i].Name] = &a[i]
	}
	return
}

func newRemoteExecutor(namespace string, opt *config.Application) *remoteExecutor {
	exe := &remoteExecutor{
		namespace:          namespace,
		opt:                opt,
		microserviceByName: microserviceArrayToMap(opt.Microservices),
	}

	return exe
}

//
// Deploy application using remote controller
//
func (exe *remoteExecutor) Execute() (err error) {
	// Get Controllers from namespace
	controllers, err := config.GetControllers(exe.namespace)

	// Do we actually have any controllers?
	if err != nil {
		util.PrintError("You must deploy a Controller to a namespace before deploying any Agents")
		return
	}

	// Did we have more than one controller?
	if len(controllers) != 1 {
		err = util.NewInternalError("Only support 1 controller per namespace")
		return
	}

	// Init remote resources
	if err = exe.init(&controllers[0]); err != nil {
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

func (exe *remoteExecutor) init(controller *config.Controller) (err error) {
	exe.client = client.New(controller.Endpoint)
	if err = exe.client.Login(client.LoginRequest{Email: controller.IofogUser.Email, Password: controller.IofogUser.Password}); err != nil {
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
	for _, route := range exe.opt.Routes {
		if _, foundFrom := exe.microserviceByName[route.From]; !foundFrom {
			return util.NewNotFoundError(fmt.Sprintf("Could not find origin microservice for the route %v", route))
		}
		if _, foundTo := exe.microserviceByName[route.To]; !foundTo {
			return util.NewNotFoundError(fmt.Sprintf("Could not find destination microservice for the route %v", route))
		}
	}

	// Validate microservice
	for _, msvc := range exe.opt.Microservices {
		if _, foundAgent := exe.agentsByName[msvc.Agent.Name]; !foundAgent {
			return util.NewNotFoundError(fmt.Sprintf("Could not find agent: %s", msvc.Agent.Name))
		}
		if _, foundCatalogItem := exe.catalogByID[msvc.Images.CatalogID]; msvc.Images.CatalogID > 0 && !foundCatalogItem {
			return util.NewNotFoundError(fmt.Sprintf("Could not find catalog item: %d", msvc.Images.CatalogID))
		}
	}

	// TODO: Check if application alredy exists
	return nil
}

func (exe *remoteExecutor) deploy() (err error) {
	defer util.SpinStop()

	// Create flow
	util.SpinStart("Creating flow")
	flow, err := exe.client.CreateFlow(exe.opt.Name, fmt.Sprintf("Flow for application: %s", exe.opt.Name))
	if err != nil {
		return
	}

	// Create microservices
	for _, msvc := range exe.opt.Microservices {
		util.SpinStart(fmt.Sprintf("Deploying microservice %s", msvc.Name))

		// TODO: Configure agent
		agent, _ := exe.agentsByName[msvc.Agent.Name]
		_, err = exe.client.UpdateAgent(&client.AgentUpdateRequest{
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
		if err != nil {
			return
		}
		// CatalogItem
		var catalogItem *client.CatalogItemInfo
		catalogImages := []client.CatalogImage{
			client.CatalogImage{ContainerImage: msvc.Images.X86, AgentTypeID: 1},
			client.CatalogImage{ContainerImage: msvc.Images.ARM, AgentTypeID: 2},
		}
		if msvc.Images.CatalogID == 0 {
			catalogItemName := fmt.Sprintf("%s_%s_catalog", exe.opt.Name, msvc.Name)
			var found bool
			catalogItem, found = exe.catalogByName[catalogItemName]
			if found == true {
				// Check if catalog item needs to be updated
				if catalogItemNeedsUpdate(catalogItem, catalogImages, msvc.Images.Registry) {
					// Delete catalog item
					if err = exe.client.DeleteCatalogItem(catalogItem.ID); err != nil {
						return
					}
					// Create new catalog item
					catalogItem, err = exe.client.CreateCatalogItem(&client.CatalogItemCreateRequest{
						Name:        catalogItemName,
						Description: fmt.Sprintf("Catalog item for %s in application %s", msvc.Name, exe.opt.Name),
						Images:      catalogImages,
						RegistryID:  msvc.Images.Registry,
					})
					if err != nil {
						return err
					}
				}
			} else {
				// Create new catalog item
				catalogItem, err = exe.client.CreateCatalogItem(&client.CatalogItemCreateRequest{
					Name:        catalogItemName,
					Description: fmt.Sprintf("Catalog item for %s in application %s", msvc.Name, exe.opt.Name),
					Images:      catalogImages,
					RegistryID:  msvc.Images.Registry,
				})
				if err != nil {
					return err
				}
			}
		} else {
			catalogItem = exe.catalogByID[msvc.Images.CatalogID]
		}

		config := ""
		if msvc.Config != nil {
			byteconfig, err := json.Marshal(msvc.Config)
			if err != nil {
				return err
			}
			config = string(byteconfig)
		}

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
		// Update msvc map with UUID
		exe.microserviceByName[msvc.Name].UUID = msvcInfo.UUID
	}

	// Create Routes
	for _, route := range exe.opt.Routes {
		fromMsvc, _ := exe.microserviceByName[route.From]
		toMsvc, _ := exe.microserviceByName[route.To]
		if err = exe.client.CreateMicroserviceRoute(fromMsvc.UUID, toMsvc.UUID); err != nil {
			return
		}
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
