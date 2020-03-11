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
	"fmt"

	jsoniter "github.com/json-iterator/go"

	apps "github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/apps"
	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func MapClientMicroserviceToDeployMicroservice(msvc *client.MicroserviceInfo, clt *client.Client) (result *apps.Microservice, err error) {
	agent, err := clt.GetAgentByID(msvc.AgentUUID)
	if err != nil {
		return
	}
	var catalogItem *client.CatalogItemInfo
	if msvc.CatalogItemID != 0 {
		catalogItem, err = clt.GetCatalogItem(msvc.CatalogItemID)
		if err != nil {
			if httpErr, ok := err.(*client.HTTPError); ok && httpErr.Code == 404 {
				catalogItem = nil
			} else {
				return nil, err
			}
		}
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

	// Map port host to agent name
	for idx, port := range msvc.Ports {
		if port.Host != "" && port.Host != iofog.VanillaRouterAgentName {
			hostAgent, err := clt.GetAgentByID(port.Host)
			var name string
			if err != nil {
				util.PrintNotify(fmt.Sprintf("Could not find agent with UUID %s\n", port.Host))
				name = "UNKNOWN_" + port.Host
			} else {
				name = hostAgent.Name
			}
			msvc.Ports[idx].Host = name
		}
	}

	jsonConfig := make(map[string]interface{})
	if err = jsoniter.Unmarshal([]byte(msvc.Config), &jsonConfig); err != nil {
		return
	}
	result = new(apps.Microservice)
	result.UUID = msvc.UUID
	result.Name = msvc.Name
	result.Agent = apps.MicroserviceAgent{
		Name: agent.Name,
		Config: apps.AgentConfiguration{
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
	var msvcImages []client.CatalogImage
	if catalogItem != nil {
		msvcImages = catalogItem.Images
	} else {
		msvcImages = msvc.Images
	}
	for _, image := range msvcImages {
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
	var registryID int
	var imgArray []client.CatalogImage
	if catalogItem != nil {
		registryID = catalogItem.RegistryID
		imgArray = catalogItem.Images
	} else {
		registryID = msvc.RegistryID
		imgArray = msvc.Images
	}
	images := apps.MicroserviceImages{
		CatalogID: msvc.CatalogItemID,
		X86:       x86Image,
		ARM:       armImage,
		Registry:  client.RegistryTypeIDRegistryTypeDict[registryID],
	}
	for _, img := range imgArray {
		if img.AgentTypeID == 1 {
			images.X86 = img.ContainerImage
		} else if img.AgentTypeID == 2 {
			images.ARM = img.ContainerImage
		}
	}
	volumes := mapVolumes(msvc.Volumes)
	envs := mapEnvs(msvc.Env)
	result.Images = &images
	result.Config = jsonConfig
	result.Container.RootHostAccess = msvc.RootHostAccess
	result.Container.Commands = msvc.Commands
	result.Container.Ports = mapPorts(msvc.Ports)
	result.Container.Volumes = &volumes
	result.Container.Env = &envs
	result.Routes = routes
	result.Flow = &application.Name
	return
}

func mapPorts(in []client.MicroservicePortMapping) (out []apps.MicroservicePortMapping) {
	for _, port := range in {
		out = append(out, apps.MicroservicePortMapping(port))
	}
	return
}

func mapVolumes(in []client.MicroserviceVolumeMapping) (out []apps.MicroserviceVolumeMapping) {
	for _, vol := range in {
		out = append(out, apps.MicroserviceVolumeMapping(vol))
	}
	return
}

func mapEnvs(in []client.MicroserviceEnvironment) (out []apps.MicroserviceEnvironment) {
	for _, env := range in {
		out = append(out, apps.MicroserviceEnvironment(env))
	}
	return
}
