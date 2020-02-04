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
	uuid        string
	namespace   string
	agentConfig *config.AgentConfiguration
}

func getAgentUpdateRequestFromAgentConfig(agentConfig config.AgentConfiguration) (request client.AgentUpdateRequest) {
	var fogTypePtr *int64
	if agentConfig.FogType != nil {
		fogType, found := config.FogTypeStringMap[*agentConfig.FogType]
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
	return
}

func CreateAgentFromConfiguration(agentConfig config.AgentConfiguration, name string, clt *client.Client) (uuid string, err error) {
	updateAgentConfigRequest := getAgentUpdateRequestFromAgentConfig(agentConfig)
	createAgentRequest := client.CreateAgentRequest{
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

	if agent.uuid != "" {
		uuid = agent.uuid
	} else {
		existingAgent, err := ctrl.GetAgentByName(agent.name)
		if err != nil {
			return "", "", err
		}
		uuid = existingAgent.UUID
	}

	// Get provisioning key
	provisionResponse, err := ctrl.GetAgentProvisionKey(uuid)
	if err != nil {
		return
	}
	key = provisionResponse.Key
	return
}
