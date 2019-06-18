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
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	pb "github.com/schollz/progressbar"
)

// Local agent uses Container exec commands
type LocalAgent struct {
	defaultAgent
	client           *LocalContainer
	localAgentConfig *LocalAgentConfig
}

func NewLocalAgent(agentConfig *LocalAgentConfig, client *LocalContainer) *LocalAgent {
	return &LocalAgent{
		defaultAgent:     defaultAgent{name: agentConfig.Name},
		localAgentConfig: agentConfig,
		client:           client,
	}
}

func (agent *LocalAgent) Bootstrap() error {
	// Nothing to do for local agent, bootstraping is done inside the image.
	return nil
}

func (agent *LocalAgent) Configure(ctrl *config.Controller, user User) (uuid string, err error) {
	pb := pb.New(100)
	defer pb.Clear()

	key, uuid, err := agent.getProvisionKey(ctrl.Endpoint, user, pb)

	// get local controller config
	ctrlConfig := NewLocalControllerConfig(ctrl.Name, make(map[string]string))

	// Use the container name as host
	ctrlContainerConfig := ctrlConfig.ContainerMap["controller"]
	controllerEndpoint := fmt.Sprintf("%s:%s", ctrlContainerConfig.ContainerName, ctrlContainerConfig.Ports[0].Host)

	// Instantiate provisioning commands
	controllerBaseURL := fmt.Sprintf("http://%s/api/v3", controllerEndpoint)
	cmds := [][]string{
		[]string{"iofog-agent", "config", "-idc", "off"},
		[]string{"iofog-agent", "config", "-a", controllerBaseURL},
		[]string{"iofog-agent", "provision", key},
	}

	// Execute commands
	for _, cmd := range cmds {
		err = agent.client.ExecuteCmd(agent.localAgentConfig.ContainerName, cmd)
		if err != nil {
			return
		}
	}

	return
}
