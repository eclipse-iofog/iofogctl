/*
 *  *******************************************************************************
 *  * Copyright (c) 2020 Edgeworx, Inc.
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
	"github.com/eclipse-iofog/iofogctl/v3/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

func MapClientMicroserviceToDeployMicroservice(msvc *client.MicroserviceInfo, clt *client.Client) (*apps.Microservice, error) {
	agent, err := clt.GetAgentByID(msvc.AgentUUID)
	if err != nil {
		return nil, err
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

	applicationName := msvc.Application
	if msvc.Application == "" {
		if msvc.FlowID > 0 {
			// Legacy
			flow, err := clt.GetFlowByID(msvc.FlowID)
			if err != nil {
				return nil, err
			}
			applicationName = flow.Name
		}
	}

	// Map port host to agent name
	for idx, port := range msvc.Ports {
		if port.Host != "" && port.Host != iofog.VanillaRouterAgentName {
			hostAgent, err := clt.GetAgentByID(port.Host)
			var name string
			if err != nil {
				util.PrintNotify(fmt.Sprintf("Could not find Agent with UUID %s\n", port.Host))
				name = "UNKNOWN_" + port.Host
			} else {
				name = hostAgent.Name
			}
			msvc.Ports[idx].Host = name
		}
	}

	return constructMicroservice(msvc, agent.Name, applicationName, catalogItem)
}

func constructMicroservice(msvcInfo *client.MicroserviceInfo, agentName, appName string, catalogItem *client.CatalogItemInfo) (msvc *apps.Microservice, err error) {
	msvc = new(apps.Microservice)
	msvc.UUID = msvcInfo.UUID
	msvc.Name = msvcInfo.Name
	msvc.Agent = apps.MicroserviceAgent{
		Name: agentName,
	}
	var armImage, x86Image string
	var msvcImages []client.CatalogImage
	if catalogItem != nil {
		msvcImages = catalogItem.Images
	} else {
		msvcImages = msvcInfo.Images
	}
	for _, image := range msvcImages {
		switch client.AgentTypeIDAgentTypeDict[image.AgentTypeID] {
		case "x86":
			x86Image = image.ContainerImage
		case "arm":
			armImage = image.ContainerImage
		default:
		}
	}
	var registryID int
	var imgArray []client.CatalogImage
	if catalogItem != nil {
		registryID = catalogItem.RegistryID
		imgArray = catalogItem.Images
	} else {
		registryID = msvcInfo.RegistryID
		imgArray = msvcInfo.Images
	}
	images := apps.MicroserviceImages{
		CatalogID: msvcInfo.CatalogItemID,
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
	volumes := mapVolumes(msvcInfo.Volumes)
	envs := mapEnvs(msvcInfo.Env)
	extraHosts := mapExtraHosts(msvcInfo.ExtraHosts)
	msvc.Images = &images
	jsonConfig := make(map[string]interface{})
	if err := jsoniter.Unmarshal([]byte(msvcInfo.Config), &jsonConfig); err != nil {
		return msvc, err
	}
	msvc.Config = jsonConfig
	msvc.Container.RootHostAccess = msvcInfo.RootHostAccess
	msvc.Container.Commands = msvcInfo.Commands
	msvc.Container.Ports = mapPorts(msvcInfo.Ports)
	msvc.Container.Volumes = &volumes
	msvc.Container.Env = &envs
	msvc.Container.ExtraHosts = &extraHosts
	msvc.Application = &appName
	return msvc, err
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

func mapExtraHosts(in []client.MicroserviceExtraHost) (out []apps.MicroserviceExtraHost) {
	for _, eH := range in {
		out = append(out, apps.MicroserviceExtraHost(eH))
	}
	return
}
