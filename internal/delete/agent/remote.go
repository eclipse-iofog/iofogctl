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

package deleteagent

import (
	"fmt"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"strings"
)

type remoteExecutor struct {
	namespace string
	name      string
}

func newRemoteExecutor(namespace, name string) *remoteExecutor {
	exe := &remoteExecutor{}
	exe.namespace = namespace
	exe.name = name
	return exe
}

func (exe *remoteExecutor) Execute() error {
	// Check the agent exists
	agent, err := config.GetAgent(exe.namespace, exe.name)
	if err != nil {
		return err
	}
	// Get Controller for the namespace
	ctrlConfigs, err := config.GetControllers(exe.namespace)
	if err != nil {
		return err
	}

	// If controller exists, deprovision the agent
	if len(ctrlConfigs) > 0 {
		// Get Controller endpoint and connect to Controller
		endpoint := ctrlConfigs[0].Endpoint
		ctrl := iofog.NewController(endpoint)

		// Log into Controller
		userConfig := ctrlConfigs[0].IofogUser
		user := iofog.LoginRequest{
			Email:    userConfig.Email,
			Password: userConfig.Password,
		}
		loginResponse, err := ctrl.Login(user)
		if err != nil {
			return err
		}
		token := loginResponse.AccessToken

		// Perform deletion of Agent through Controller
		if err = ctrl.DeleteAgent(agent.UUID, token); err != nil {
			if !strings.Contains(err.Error(), "NotFoundError") {
				return err
			}
		}
	}

	// Update configuration
	if err = config.DeleteAgent(exe.namespace, exe.name); err != nil {
		return err
	}

	fmt.Printf("\nAgent %s/%s successfully deleted.\n", exe.namespace, exe.name)

	return config.Flush()
}
