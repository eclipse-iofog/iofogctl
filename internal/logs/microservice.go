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

package logs

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/eclipse-iofog/iofogctl/internal"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
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
	agent, image, err := getAgentAndImage(ms.namespace, ms.name)
	if err != nil {
		return err
	}

	// Verify we can SSH into the Agent
	if agent.SSH.User == "" || agent.Host == "" || agent.SSH.KeyFile == "" {
		return util.NewError("Cannot get logs for microservice on Agent " + agent.Name + " because SSH details are not available")
	}

	// SSH into the Agent and get the logs
	ssh := util.NewSecureShellClient(agent.SSH.User, agent.Host, agent.SSH.KeyFile)
	if err = ssh.Connect(); err != nil {
		return err
	}

	// Notify the user of the containers that are up
	out, err := ms.runDockerCommand(fmt.Sprintf("docker ps | grep %s", image), ssh)
	if err != nil {
		return err
	}
	msg := "Retrieving logs for the first container in this list\n" + out.String()
	util.PrintInfo(msg)

	// Execute the command
	logFile := util.After(image, "/")
	cmd := fmt.Sprintf("docker ps | grep %s | awk 'FNR == 1 {print $1}' | xargs docker logs 2> /tmp/%s.logs", image, logFile)
	out, err = ms.runDockerCommand(cmd, ssh)
	if err != nil {
		return err
	}

	// Output stdout of the logs
	fmt.Println(out.String())

	// Execute command to print stderr
	cmd = fmt.Sprintf("cat /tmp/%s.logs", logFile)
	out, err = ssh.Run(cmd)
	if err != nil {
		return err
	}

	// Output stderr of the logs
	fmt.Println(out.String())

	return nil
}

func (ms *remoteMicroserviceExecutor) runDockerCommand(cmd string, ssh *util.SecureShellClient) (stdout bytes.Buffer, err error) {
	stdout, err = ssh.Run(cmd)
	if err != nil {
		if !strings.Contains(err.Error(), "permission denied") {
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

func getAgentAndImage(namespace, msvcName string) (agent config.Agent, image string, err error) {
	ctrlClient, err := internal.NewControllerClient(namespace)
	if err != nil {
		return
	}

	// Get microservice details from Controller
	msvc, err := ctrlClient.GetMicroserviceByName(msvcName)
	if err != nil {
		return
	}

	// Images must exist
	if len(msvc.Images) == 0 {
		err = util.NewError("Microservice " + msvcName + " does not have any images")
		return
	}
	image = msvc.Images[0].ContainerImage

	// Get Agent running the microservice
	agentResponse, err := ctrlClient.GetAgentByID(msvc.AgentUUID)
	if err != nil {
		return
	}
	agent, err = config.GetAgent(namespace, agentResponse.Name)
	if err != nil {
		return
	}
	return
}
