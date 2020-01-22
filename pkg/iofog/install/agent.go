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

package install

import (
	"github.com/eclipse-iofog/iofog-go-sdk/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
)

type Agent interface {
	Bootstrap() error
	getProvisionKey(string, IofogUser) (string, string, error)
	Configure(*config.Controller, IofogUser) (string, error)
}

// defaultAgent implements commong behavior
type defaultAgent struct {
	name        string
	namespace   string
	agentConfig *config.AgentConfiguration
}

func getAgentUpdateRequestFromAgentConfig(agentConfig config.AgentConfiguration) (request client.AgentUpdateRequest) {
	fogType, found := config.FogTypeStringMap[agentConfig.FogType]
	if !found {
		fogType = 0
	}
	request.Location = agentConfig.Location
	request.Latitude = agentConfig.Latitude
	request.Longitude = agentConfig.Longitude
	request.Description = agentConfig.Description
	request.Name = agentConfig.Name
	request.FogType = fogType
	request.AgentConfiguration = agentConfig.AgentConfiguration
	return
}

func UpdateAgentConfiguration(agentConfig *config.AgentConfiguration, uuid string, clt *client.Client) (err error) {
	if agentConfig != nil {
		updateAgentConfigRequest := getAgentUpdateRequestFromAgentConfig(*agentConfig)
		updateAgentConfigRequest.UUID = uuid

		if _, err = clt.UpdateAgent(&updateAgentConfigRequest); err != nil {
			return
		}
	}
	return nil
}

func (agent *defaultAgent) getProvisionKey(controllerEndpoint string, user IofogUser) (key string, uuid string, err error) {
	// Connect to controller
	ctrl := client.New(controllerEndpoint)

	// Log in
	Verbose("Accessing Controller to generate Provisioning Key")
	loginRequest := client.LoginRequest{
		Email:    user.Email,
		Password: user.Password,
	}
	if err = ctrl.Login(loginRequest); err != nil {
		return
	}

	// If the agent already exists, re-use the UUID
	agentList, err := ctrl.ListAgents()
	if err != nil {
		return
	}
	for _, existingAgent := range agentList.Agents {
		if existingAgent.Name == agent.name {
			uuid = existingAgent.UUID
			break
		}
	}

	// Create agent if necessary
	if uuid == "" {
		var updateAgentRequest client.AgentUpdateRequest
		if agent.agentConfig != nil {
			updateAgentRequest = getAgentUpdateRequestFromAgentConfig(*agent.agentConfig)
		}
		updateAgentRequest.Name = agent.name
		createRequest := client.CreateAgentRequest{
			AgentUpdateRequest: updateAgentRequest,
		}
		var createResponse client.CreateAgentResponse
		createResponse, err = ctrl.CreateAgent(createRequest)
		if err != nil {
			return
		}
		uuid = createResponse.UUID
	}

	// Get provisioning key
	provisionResponse, err := ctrl.GetAgentProvisionKey(uuid)
	if err != nil {
		return
	}
	key = provisionResponse.Key
	return
}
