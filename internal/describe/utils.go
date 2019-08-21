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

package describe

import (
	"encoding/json"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/client"
)

func MapClientMicroserviceToConfigMicroservice(msvc *client.MicroserviceInfo, clt *client.Client) (result *config.Microservice, err error) {
	agent, err := clt.GetAgentByID(msvc.AgentUUID)
	if err != nil {
		return
	}
	catalogItem, err := clt.GetCatalogItem(msvc.CatalogItemID)
	if err != nil {
		return
	}
	application, err := clt.GetFlowByID(msvc.FlowID)
	if err != nil {
		return
	}

	routes := []string{}

	for _, msvcUUID := range msvc.Routes {
		destMsvc, err := clt.GetMicroserviceByID(msvcUUID)
		if err != nil {
			return nil, err
		}
		routes = append(routes, destMsvc.Name)
	}

	jsonConfig := make(map[string]interface{})
	if err = json.Unmarshal([]byte(msvc.Config), &jsonConfig); err != nil {
		return
	}
	result = new(config.Microservice)
	result.UUID = msvc.UUID
	result.Name = msvc.Name
	result.Agent = config.MicroserviceAgent{
		Name: agent.Name,
		Config: client.AgentConfiguration{
			DockerURL:                 &agent.DockerURL,
			DiskLimit:                 &agent.DiskLimit,
			DiskDirectory:             &agent.DiskDirectory,
			MemoryLimit:               &agent.MemoryLimit,
			CPULimit:                  &agent.CPULimit,
			LogLimit:                  &agent.LogLimit,
			LogDirectory:              &agent.LogDirectory,
			LogFileCount:              &agent.LogFileCount,
			StatusFrequency:           &agent.StatusFrequency,
			ChangeFrequency:           &agent.ChangeFrequency,
			DeviceScanFrequency:       &agent.DeviceScanFrequency,
			BluetoothEnabled:          &agent.BluetoothEnabled,
			WatchdogEnabled:           &agent.WatchdogEnabled,
			AbstractedHardwareEnabled: &agent.AbstractedHardwareEnabled,
		},
	}
	var armImage, x86Image string
	for _, image := range catalogItem.Images {
		switch client.AgentTypeIDAgentTypeDict[image.AgentTypeID] {
		case "x86":
			x86Image = image.ContainerImage
			break
		case "arm":
			armImage = image.ContainerImage
			break
		default:
			break
		}
	}
	images := config.MicroserviceImages{
		CatalogID: catalogItem.ID,
		X86:       x86Image,
		ARM:       armImage,
		Registry:  client.RegistryTypeIDRegistryTypeDict[catalogItem.RegistryID],
	}
	for _, img := range catalogItem.Images {
		if img.AgentTypeID == 1 {
			images.X86 = img.ContainerImage
		} else if img.AgentTypeID == 2 {
			images.ARM = img.ContainerImage
		}
	}
	result.Images = images
	result.Config = jsonConfig
	result.RootHostAccess = msvc.RootHostAccess
	result.Ports = msvc.Ports
	result.Volumes = msvc.Volumes
	result.Routes = routes
	result.Env = msvc.Env
	result.Flow = &application.Name
	return
}
