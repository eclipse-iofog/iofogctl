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

	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v2/internal"
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
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
	agent, agentInfo, msvc, err := getAgentAndMicroservice(ms.namespace, ms.name)
	if err != nil {
		return err
	}

	if msvc.Status.Status != "RUNNING" {
		return util.NewError("The microservice is not currently running")
	}

	// Local
	if util.IsLocalHost(agent.Host) {
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
	}

	// Verify we can SSH into the Agent
	if agent.SSH.User == "" || agent.Host == "" || agent.SSH.KeyFile == "" {
		return util.NewError("Cannot get logs for microservice on Agent " + agent.Name + " because SSH details are not available")
	}

	// SSH into the Agent and get the logs
	ssh := util.NewSecureShellClient(agent.SSH.User, agent.Host, agent.SSH.KeyFile)
	ssh.SetPort(agent.SSH.Port)
	if err = ssh.Connect(); err != nil {
		return err
	}

	// Notify the user of the containers that are up
	var image string
	for _, img := range msvc.Images {
		if img.AgentTypeID == agentInfo.FogType {
			image = img.ContainerImage
		}
	}
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

func getAgentAndMicroservice(namespace, msvcName string) (agent rsc.Agent, agentInfo client.AgentInfo, msvc client.MicroserviceInfo, err error) {
	ctrlClient, err := internal.NewControllerClient(namespace)
	if err != nil {
		return
	}

	// Get microservice details from Controller
	msvcPtr, err := ctrlClient.GetMicroserviceByName(msvcName)
	if err != nil {
		return
	}

	msvc = *msvcPtr

	// Images must exist
	if len(msvc.Images) == 0 {
		err = util.NewError("Microservice " + msvcName + " does not have any images")
		return
	}

	// Get Agent running the microservice
	agentResponse, err := ctrlClient.GetAgentByID(msvc.AgentUUID)
	if err != nil {
		return
	}
	agentInfo = *agentResponse
	agent, err = config.GetAgent(namespace, agentResponse.Name)
	if err != nil {
		return
	}
	return
}
