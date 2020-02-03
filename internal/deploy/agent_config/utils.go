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

package deployagentconfig

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type RouterMode string

const (
	EdgeRouter     RouterMode = "edge"
	InteriorRouter RouterMode = "interior"
	NoneRouter     RouterMode = "none"
)

func Validate(config config.AgentConfiguration) error {
	var routerMode RouterMode
	if config.RouterConfig.RouterMode != nil {
		routerMode = RouterMode(*config.RouterConfig.RouterMode)
	} else {
		routerMode = EdgeRouter
	}

	if routerMode != EdgeRouter && routerMode != InteriorRouter && routerMode != NoneRouter {
		return util.NewInputError(fmt.Sprintf("Agent config %s validation failed. RouterMode has to be one of edge, interior, none. Default is: edge", config.Name))
	}
	if routerMode != NoneRouter && config.NetworkRouter != nil {
		return util.NewInputError(fmt.Sprintf("Agent config %s validation failed. Cannot have a network if routerMode is different from none. Default router mode is edge", config.Name))
	}
	if routerMode == NoneRouter && config.UpstreamRouters != nil && len(*config.UpstreamRouters) > 0 {
		return util.NewInputError(fmt.Sprintf("Agent config %s validation failed. Cannot have a upstreamRouters if routerMode is none", config.Name))
	}
	if routerMode != InteriorRouter && config.RouterConfig.EdgeRouterPort != nil || config.RouterConfig.InterRouterPort != nil {
		return util.NewInputError(fmt.Sprintf("Agent config %s validation failed. Cannot have a edgeRouterPort of InterRouterPort if routerMode is different from interior. Default router mode is edge", config.Name))
	}

	return nil
}

func findAgentUuidInList(list []client.AgentInfo, name string) (uuid string, err error) {
	if name == iofog.VanillaRouterAgentName {
		return name, nil
	}
	for _, agent := range list {
		if agent.Name == name {
			return agent.UUID, nil
		}
	}
	return "", util.NewNotFoundError(fmt.Sprintf("Could not find router: %s\n", name))
}

func ProcessAgentNames(config config.AgentConfiguration, clt *client.Client) (config.AgentConfiguration, error) {
	agentList, err := clt.ListAgents()
	if err != nil {
		return config, err
	}

	if config.UpstreamRouters != nil {
		upstreamRoutersUUID := []string{}
		for _, agentName := range *config.UpstreamRouters {
			uuid, err := findAgentUuidInList(agentList.Agents, agentName)
			if err != nil {
				return config, err
			}
			upstreamRoutersUUID = append(upstreamRoutersUUID, uuid)
		}
		config.UpstreamRouters = &upstreamRoutersUUID
	}

	if config.NetworkRouter != nil {
		uuid, err := findAgentUuidInList(agentList.Agents, *config.NetworkRouter)
		if err != nil {
			return config, err
		}
		config.NetworkRouter = &uuid
	}
	return config, nil
}
