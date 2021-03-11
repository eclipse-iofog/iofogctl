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

package deployagentconfig

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

type RouterMode string

const (
	EdgeRouter     RouterMode = "edge"
	InteriorRouter RouterMode = "interior"
	NoneRouter     RouterMode = "none"
)

func getRouterMode(config *rsc.AgentConfiguration) RouterMode {
	if config.RouterConfig.RouterMode != nil {
		return RouterMode(*config.RouterConfig.RouterMode)
	}
	return EdgeRouter
}

func Validate(config *rsc.AgentConfiguration) error {
	routerMode := getRouterMode(config)

	if routerMode != EdgeRouter && routerMode != InteriorRouter && routerMode != NoneRouter {
		msg := "agent config %s validation failed. RouterMode has to be one of edge, interior, none. Default is: edge"
		return util.NewInputError(fmt.Sprintf(msg, config.Name))
	}
	if routerMode != NoneRouter && config.NetworkRouter != nil {
		msg := "agent config %s validation failed. Cannot have a network if routerMode is different from none. Current router mode is: %s"
		return util.NewInputError(fmt.Sprintf(msg, config.Name, routerMode))
	}
	if routerMode == NoneRouter && config.UpstreamRouters != nil && len(*config.UpstreamRouters) > 0 {
		msg := "agent config %s validation failed. Cannot have a upstreamRouters if routerMode is none"
		return util.NewInputError(fmt.Sprintf(msg, config.Name))
	}
	if routerMode != InteriorRouter && (config.RouterConfig.EdgeRouterPort != nil || config.RouterConfig.InterRouterPort != nil) {
		msg := "agent config %s validation failed. Cannot have an edgeRouterPort or interRouterPort if routerMode is different from interior. Current router mode is: %s"
		return util.NewInputError(fmt.Sprintf(msg, config.Name, routerMode))
	}

	return nil
}

func findAgentUUIDInList(list []client.AgentInfo, name string) (uuid string, err error) {
	if name == iofog.VanillaRouterAgentName {
		return name, nil
	}
	for idx := range list {
		agent := &list[idx]
		if agent.Name == name {
			return agent.UUID, nil
		}
	}
	return "", util.NewNotFoundError(fmt.Sprintf("Could not find router: %s\n", name))
}

// Process update the config to translate agent names into uuids, and sets the host value if needed
func Process(agentConfig *rsc.AgentConfiguration, name, agentIP string, otherAgents []client.AgentInfo) error {
	routerMode := getRouterMode(agentConfig)

	if agentConfig.UpstreamRouters != nil {
		upstreamRoutersUUID := []string{}
		for _, agentName := range *agentConfig.UpstreamRouters {
			uuid, err := findAgentUUIDInList(otherAgents, agentName)
			if err != nil {
				return err
			}
			upstreamRoutersUUID = append(upstreamRoutersUUID, uuid)
		}
		agentConfig.UpstreamRouters = &upstreamRoutersUUID
	}

	if agentConfig.NetworkRouter != nil {
		uuid, err := findAgentUUIDInList(otherAgents, *agentConfig.NetworkRouter)
		if err != nil {
			return err
		}
		agentConfig.NetworkRouter = &uuid
	}

	if routerMode != NoneRouter && agentConfig.Host == nil {
		agentConfig.Host = &agentIP
	}

	return nil
}

func getAgentUpdateRequestFromAgentConfig(agentConfig *rsc.AgentConfiguration, tags *[]string) (request client.AgentUpdateRequest) {
	var fogTypePtr *int64
	if agentConfig.FogType != nil {
		fogType, found := rsc.FogTypeStringMap[*agentConfig.FogType]
		if !found {
			fogType = 0
		}
		fogTypePtr = &fogType
	}
	request.Location = agentConfig.Location
	request.Latitude = agentConfig.Latitude
	request.Longitude = agentConfig.Longitude
	request.Description = agentConfig.Description
	request.Name = agentConfig.Name
	request.FogType = fogTypePtr
	request.AgentConfiguration = agentConfig.AgentConfiguration
	request.Tags = tags
	return
}

func createAgentFromConfiguration(agentConfig *rsc.AgentConfiguration, tags *[]string, name string, clt *client.Client) (uuid string, err error) {
	updateAgentConfigRequest := getAgentUpdateRequestFromAgentConfig(agentConfig, tags)
	createAgentRequest := &client.CreateAgentRequest{
		AgentUpdateRequest: updateAgentConfigRequest,
	}
	if createAgentRequest.AgentUpdateRequest.Name == "" {
		createAgentRequest.AgentUpdateRequest.Name = name
	}
	if createAgentRequest.AgentUpdateRequest.FogType == nil {
		fogType := int64(0)
		createAgentRequest.AgentUpdateRequest.FogType = &fogType
	}
	agent, err := clt.CreateAgent(createAgentRequest)
	if err != nil {
		return "", err
	}
	return agent.UUID, nil
}

func updateAgentConfiguration(agentConfig *rsc.AgentConfiguration, tags *[]string, uuid string, clt *client.Client) (err error) {
	if agentConfig != nil {
		updateAgentConfigRequest := getAgentUpdateRequestFromAgentConfig(agentConfig, tags)
		updateAgentConfigRequest.UUID = uuid

		if _, err = clt.UpdateAgent(&updateAgentConfigRequest); err != nil {
			return
		}
	}
	return nil
}
