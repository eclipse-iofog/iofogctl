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
		createRequest := client.CreateAgentRequest{
			Name:    agent.name,
			FogType: 0,
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
