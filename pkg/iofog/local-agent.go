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

func (agent *LocalAgent) Configure(controllerEndpoint string, user User) (uuid string, err error) {
	pb := pb.New(100)
	defer pb.Clear()

	key, uuid, err := agent.getProvisionKey(controllerEndpoint, user, pb)

	// Instantiate provisioning commands
	controllerBaseURL := fmt.Sprintf("http://%s/api/v3", controllerEndpoint)
	cmds := []command{
		{fmt.Sprintf("sh -c 'iofog-agent config -a %s'", controllerBaseURL), 10},
		{fmt.Sprintf("sh -c 'iofog-agent provision %s'", key), 10},
	}

	// Execute commands
	for _, cmd := range cmds {
		containerCmd := []string{cmd.cmd}
		err = agent.client.ExecuteCmd(agent.localAgentConfig.ContainerName, containerCmd)
		if err != nil {
			return
		}
	}

	return
}
