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
	pb "github.com/schollz/progressbar"
)

type command struct {
	cmd     string
	pbSlice int
}

type Agent interface {
	Bootstrap() error
	getProvisionKey(string, User, *pb.ProgressBar) (string, string, error)
	Configure(string, User) (string, error)
}

// defaultAgent implements commong behavior
type defaultAgent struct {
	name string
}

func (agent *defaultAgent) getProvisionKey(controllerEndpoint string, user User, pb *pb.ProgressBar) (key string, uuid string, err error) {
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
	pb.Add(20)

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
	pb.Add(20)

	// Get provisioning key
	provisionResponse, err := ctrl.GetAgentProvisionKey(uuid, token)
	if err != nil {
		return
	}
	pb.Add(20)
	key = provisionResponse.Key
	return
}
