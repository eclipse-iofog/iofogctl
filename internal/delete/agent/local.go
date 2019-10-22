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
	"strings"

	"github.com/eclipse-iofog/iofogctl/pkg/util"

	"github.com/eclipse-iofog/iofog-go-sdk/pkg/client"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
)

type localExecutor struct {
	namespace        string
	name             string
	client           *install.LocalContainer
	localAgentConfig *install.LocalAgentConfig
}

func newLocalExecutor(namespace, name string, client *install.LocalContainer) *localExecutor {
	ctrlConfig := install.NewLocalControllerConfig(make(map[string]string), install.Credentials{})
	exe := &localExecutor{
		namespace:        namespace,
		name:             name,
		client:           client,
		localAgentConfig: install.NewLocalAgentConfig(name, "", ctrlConfig, install.Credentials{}),
	}
	return exe
}

func (exe *localExecutor) GetName() string {
	return exe.name
}

func (exe *localExecutor) Execute() error {
	// Get Control Plane config details
	controlPlane, err := config.GetControlPlane(exe.namespace)
	if err != nil {
		return err
	}

	iofogClient, err := client.NewAndLogin(controlPlane.Controllers[0].Endpoint, controlPlane.IofogUser.Email, controlPlane.IofogUser.Password)
	if err != nil {
		return err
	}

	// Get agent UUID
	agentList, err := iofogClient.ListAgents()
	if err != nil {
		return err
	}
	var agentUUID string
	for _, agent := range agentList.Agents {
		if agent.Name == exe.name {
			agentUUID = agent.UUID
			break
		}
	}

	// Get list of microservices
	microservicesList, err := iofogClient.GetAllMicroservices()
	if err != nil {
		return err
	}

	// Clean agent container
	if errClean := exe.client.CleanContainer(exe.localAgentConfig.ContainerName); errClean != nil {
		util.PrintNotify(fmt.Sprintf("Could not clean Agent container: %v", errClean))
	}

	// Clean microservices
	for _, msvc := range microservicesList.Microservices {
		if agentUUID == msvc.AgentUUID {
			if errClean := exe.client.CleanContainer(fmt.Sprintf("iofog_%s", msvc.UUID)); errClean != nil {
				util.PrintNotify(fmt.Sprintf("Could not clean Microservice container: %v", errClean))
			}
		}
	}

	// Perform deletion of Agent through Controller
	if err = iofogClient.DeleteAgent(agentUUID); err != nil {
		if !strings.Contains(err.Error(), "NotFoundError") {
			return err
		}
	}

	// Update config
	if err := config.DeleteAgent(exe.namespace, exe.name); err != nil {
		return err
	}

	return nil
}
