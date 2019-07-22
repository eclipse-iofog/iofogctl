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
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Agent interface {
	Bootstrap() error
	getProvisionKey(string, client.User) (string, string, error)
	Configure(*config.Controller, client.User) (string, error)
}

// defaultAgent implements commong behavior
type defaultAgent struct {
	name      string
	namespace string
}

func (agent *defaultAgent) getProvisionKey(controllerEndpoint string, user client.User) (key string, uuid string, err error) {
	// Connect to controller
	ctrl := client.New(controllerEndpoint)

	// Log in
	loginRequest := client.LoginRequest{
		Email:    user.Email,
		Password: user.Password,
	}
	loginResponse, err := ctrl.Login(loginRequest)
	if err != nil {
		return
	}
	token := loginResponse.AccessToken

	// Check the agent name is unique
	var agentList client.ListAgentsResponse
	agentList, err = ctrl.ListAgents(token)
	if err != nil {
		return
	}
	for _, existingAgent := range agentList.Agents {
		if existingAgent.Name == agent.name {
			err = util.NewConflictError("Agent name " + agent.name + " is already registered with the Controller")
			return
		}
	}

	// Create agent
	createRequest := client.CreateAgentRequest{
		Name:    agent.name,
		FogType: 0,
	}
	createResponse, err := ctrl.CreateAgent(createRequest, token)
	if err != nil {
		return
	}
	uuid = createResponse.UUID

	// Get provisioning key
	provisionResponse, err := ctrl.GetAgentProvisionKey(uuid, token)
	if err != nil {
		return
	}
	key = provisionResponse.Key
	return
}
