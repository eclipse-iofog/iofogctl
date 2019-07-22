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

package connect

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	client "github.com/eclipse-iofog/iofogctl/pkg/iofog/client"
)

func connect(opt *Options, endpoint string) error {
	// Connect to Controller
	ctrl := client.NewController(endpoint)

	// Get sanitized endpoint
	endpoint = ctrl.GetEndpoint()

	// Login user
	loginRequest := client.LoginRequest{
		Email:    opt.Email,
		Password: opt.Password,
	}
	loginResponse, err := ctrl.Login(loginRequest)
	if err != nil {
		return err
	}
	token := loginResponse.AccessToken

	// Get Agents
	listAgentsResponse, err := ctrl.ListAgents(token)
	if err != nil {
		return err
	}

	// Update Agents config
	for _, agent := range listAgentsResponse.Agents {
		agentConfig := config.Agent{
			Name: agent.Name,
			UUID: agent.UUID,
			Host: agent.IPAddress,
		}
		err = config.AddAgent(opt.Namespace, agentConfig)
		if err != nil {
			return err
		}
	}

	// Update Controller config
	ctrlConfig := config.Controller{
		Name:     opt.Name,
		Endpoint: endpoint,
		IofogUser: config.IofogUser{
			Email:    opt.Email,
			Password: opt.Password,
		},
		KubeConfig: opt.KubeFile,
	}
	err = config.AddController(opt.Namespace, ctrlConfig)
	if err != nil {
		return err
	}

	return nil
}
