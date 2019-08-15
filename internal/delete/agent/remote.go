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
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/client"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
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

func (exe *remoteExecutor) GetName() string {
	return exe.name
}

func (exe *remoteExecutor) Execute() error {
	// Check the agent exists
	agent, err := config.GetAgent(exe.namespace, exe.name)
	if err != nil {
		return err
	}
	// Get Control Plane for the namespace
	controlPlane, err := config.GetControlPlane(exe.namespace)
	if err != nil {
		return err
	}

	// If controller exists, deprovision the agent
	if len(controlPlane.Controllers) > 0 {
		// TODO: change [0] with controlPlane variable
		// Get Controller endpoint and connect to Controller
		endpoint := controlPlane.Controllers[0].Endpoint
		ctrl := client.New(endpoint)

		// Log into Controller
		user := client.LoginRequest{
			Email:    controlPlane.IofogUser.Email,
			Password: controlPlane.IofogUser.Password,
		}
		if err := ctrl.Login(user); err != nil {
			return err
		}

		// Perform deletion of Agent through Controller
		if err = ctrl.DeleteAgent(agent.UUID); err != nil {
			if !strings.Contains(err.Error(), "NotFoundError") {
				return err
			}
		}
	}

	// Stop the Agent process on remote server
	if agent.Host == "" || agent.User == "" || agent.KeyFile == "" || agent.Port == 0 {
		util.PrintNotify("Cannot stop Agent process on remote server because SSH details for server are not available")
	} else {
		sshAgent := install.NewRemoteAgent(agent.User, agent.Host, agent.Port, agent.KeyFile, agent.Name)
		if err = sshAgent.Stop(); err != nil {
			util.PrintNotify(fmt.Sprintf("Failed to stop Agent process on remote server: %s", err.Error()))
		}
	}

	return nil
}
