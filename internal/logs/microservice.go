/*
 *  *******************************************************************************
 *  * Copyright (c) 2020 Edgeworx, Inc.
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package logs

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/v3/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

type remoteMicroserviceExecutor struct {
	namespace string
	name      string
}

func newRemoteMicroserviceExecutor(namespace, name string) *remoteMicroserviceExecutor {
	m := &remoteMicroserviceExecutor{}
	m.namespace = namespace
	m.name = name
	return m
}

func (ms *remoteMicroserviceExecutor) GetName() string {
	return ms.name
}

func (ms *remoteMicroserviceExecutor) Execute() error {
	// Get image name of the microservice and details of the Agent its deployed on
	baseAgent, msvc, err := getAgentAndMicroservice(ms.namespace, ms.name)
	if err != nil {
		return err
	}

	if msvc.Status.Status != "RUNNING" {
		return util.NewError("The microservice is not currently running")
	}

	switch agent := baseAgent.(type) {
	case *rsc.LocalAgent:
		lc, err := install.NewLocalContainerClient()
		if err != nil {
			return err
		}
		containerName := "iofog_" + msvc.UUID
		stdout, stderr, err := lc.GetLogsByName(containerName)
		if err != nil {
			return err
		}

		printContainerLogs(stdout, stderr)

		return nil
	case *rsc.RemoteAgent:
		// Verify we can SSH into the Agent
		if err := agent.ValidateSSH(); err != nil {
			return err
		}

		// SSH into the Agent and get the logs
		ssh, err := util.NewSecureShellClient(agent.SSH.User, agent.Host, agent.SSH.KeyFile)
		if err != nil {
			return err
		}
		ssh.SetPort(agent.SSH.Port)
		if err := ssh.Connect(); err != nil {
			return err
		}

		// Notify the user of the containers that are up
		containerName := "iofog_" + msvc.UUID
		out, err := ms.runDockerCommand(fmt.Sprintf("docker ps | grep %s", containerName), ssh)
		if err != nil {
			return err
		}

		// Execute the command
		cmd := fmt.Sprintf("docker ps | grep %s | awk 'FNR == 1 {print $1}' | xargs docker logs", containerName)
		out, err = ms.runDockerCommand(cmd, ssh)
		if err != nil {
			return err
		}

		// Output stdout of the logs
		fmt.Println(out.String())
	}

	return nil
}

func (ms *remoteMicroserviceExecutor) runDockerCommand(cmd string, ssh *util.SecureShellClient) (stdout bytes.Buffer, err error) {
	stdout, err = ssh.Run(cmd)
	if err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "permission denied") {
			return
		}
		// Retry with sudo
		cmd = strings.Replace(cmd, "docker", "sudo docker", -1)

		stdout, err = ssh.Run(cmd)
		if err != nil {
			return
		}
	}
	return
}

func getAgentAndMicroservice(namespace, msvcFQName string) (agent rsc.Agent, msvc client.MicroserviceInfo, err error) {
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return
	}

	ctrlClient, err := clientutil.NewControllerClient(namespace)
	if err != nil {
		return
	}

	appName, msvcName, err := clientutil.ParseFQName(msvcFQName, "Microservice")
	if err != nil {
		return agent, msvc, err
	}

	// Get microservice details from Controller
	msvcPtr, err := ctrlClient.GetMicroserviceByName(appName, msvcName)
	if err != nil {
		return
	}

	msvc = *msvcPtr

	// Get Agent running the microservice
	agentResponse, err := ctrlClient.GetAgentByID(msvc.AgentUUID)
	if err != nil {
		return
	}
	agent, err = ns.GetAgent(agentResponse.Name)
	if err != nil {
		return
	}
	return agent, msvc, nil
}
