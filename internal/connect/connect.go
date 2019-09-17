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
	client "github.com/eclipse-iofog/iofog-go-sdk/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
)

func connect(opt *Options, endpoint string) error {
	// Connect to Controller
	ctrl := client.New(endpoint)

	// Get sanitized endpoint
	endpoint = ctrl.GetEndpoint()

	// Login user
	loginRequest := client.LoginRequest{
		Email:    opt.Email,
		Password: opt.Password,
	}
	if err := ctrl.Login(loginRequest); err != nil {
		return err
	}

	// Get Connectors
	listConnectorsResponse, err := ctrl.ListConnectors()
	if err != nil {
		return err
	}

	// Update Connectors config
	for _, connector := range listConnectorsResponse.Connectors {
		connectorConfig := config.Connector{
			Name: connector.Name,
			Host: connector.IP,
		}
		if err = config.AddConnector(opt.Namespace, connectorConfig); err != nil {
			return err
		}
	}

	// Get Agents
	listAgentsResponse, err := ctrl.ListAgents()
	if err != nil {
		return err
	}

	// Update Agents config
	for _, agent := range listAgentsResponse.Agents {
		agentConfig := config.Agent{
			Name: agent.Name,
			UUID: agent.UUID,
			Host: agent.IPAddressExternal,
		}
		if err = config.AddAgent(opt.Namespace, agentConfig); err != nil {
			return err
		}
	}

	// TODO: We want to be able to connect to all Controllers in a namespace, but we can't right now
	// Update Controller config
	controlPlane := config.ControlPlane{
		IofogUser: config.IofogUser{
			Email:    opt.Email,
			Password: opt.Password,
		},
		Controllers: []config.Controller{
			{
				Name:       opt.Name,
				Endpoint:   endpoint,
				KubeConfig: opt.KubeFile,
			},
		},
	}
	err = config.UpdateControlPlane(opt.Namespace, controlPlane)
	if err != nil {
		return err
	}

	return nil
}
