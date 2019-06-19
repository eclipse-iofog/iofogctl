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

package iofog

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	pb "github.com/schollz/progressbar"
)

type command struct {
	cmd     string
	pbSlice int
}

type Agent interface {
	Bootstrap() error
	getProvisionKey(string, User, *pb.ProgressBar) (string, string, error)
	Configure(*config.Controller, User) (string, error)
}

// defaultAgent implements commong behavior
type defaultAgent struct {
	name      string
	namespace string
	pb        *pb.ProgressBar
}

func (agent *defaultAgent) getProvisionKey(controllerEndpoint string, user User) (key string, uuid string, err error) {
	// Connect to controller
	ctrl := NewController(controllerEndpoint)

	// Log in
	loginRequest := LoginRequest{
		Email:    user.Email,
		Password: user.Password,
	}
	loginResponse, err := ctrl.Login(loginRequest)
	if err != nil {
		return
	}
	token := loginResponse.AccessToken
	agent.pb.Add(20)

	// Create agent
	createRequest := CreateAgentRequest{
		Name:    agent.name,
		FogType: 0,
	}
	createResponse, err := ctrl.CreateAgent(createRequest, token)
	if err != nil {
		return
	}
	uuid = createResponse.UUID
	agent.pb.Add(20)

	// Get provisioning key
	provisionResponse, err := ctrl.GetAgentProvisionKey(uuid, token)
	if err != nil {
		return
	}
	agent.pb.Add(20)
	key = provisionResponse.Key
	return
}
