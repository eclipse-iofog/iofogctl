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
	"fmt"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

// LocalAgent uses Container exec commands
type LocalAgent struct {
	defaultAgent
	client           *LocalContainer
	localAgentConfig *LocalAgentConfig
}

func NewLocalAgent(localAgentConfig *LocalAgentConfig, client *LocalContainer) *LocalAgent {
	return &LocalAgent{
		defaultAgent:     defaultAgent{name: localAgentConfig.Name},
		localAgentConfig: localAgentConfig,
		client:           client,
	}
}

func (agent *LocalAgent) Bootstrap() error {
	// Nothing to do for local agent, bootstraping is done inside the image.
	return nil
}

func (agent *LocalAgent) Configure(controllerEndpoint string, user IofogUser) (uuid string, err error) {
	controllerEndpoint, err = agent.client.GetLocalControllerEndpoint()
	if err != nil {
		return "", err
	}
	key, uuid, err := agent.getProvisionKey("localhost", user)
	if err != nil {
		return "", err
	}

	// Instantiate provisioning commands
	controllerBaseURL := fmt.Sprintf("http://%s/api/v3", controllerEndpoint)
	cmds := [][]string{
		{"iofog-agent", "config", "-idc", "off"},
		{"iofog-agent", "config", "-a", controllerBaseURL},
		{"iofog-agent", "provision", key},
		{"iofog-agent", "config", "-sf", "10", "-cf", "10"},
	}

	// Execute commands
	for _, cmd := range cmds {
		result, err := agent.client.ExecuteCmd(agent.localAgentConfig.ContainerName, cmd)
		if result.ExitCode != 0 {
			return "", util.NewError(fmt.Sprintf("Command: %v failed with exit code %d\nStdout: %s\n Stderr: %s\n", cmd, result.ExitCode, result.StdOut, result.StdErr))
		}
		if err != nil {
			return "", err
		}
	}

	return
}
