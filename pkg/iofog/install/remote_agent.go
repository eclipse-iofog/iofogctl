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
	"os"
	"strings"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

// Remote agent uses SSH
type RemoteAgent struct {
	defaultAgent
	ssh *util.SecureShellClient
}

func NewRemoteAgent(user, host string, port int, privKeyFilename, agentName string) *RemoteAgent {
	ssh := util.NewSecureShellClient(user, host, privKeyFilename)
	ssh.SetPort(port)
	return &RemoteAgent{
		defaultAgent: defaultAgent{name: agentName},
		ssh:          ssh,
	}
}

func (agent *RemoteAgent) Bootstrap() error {
	defer util.SpinStop()
	util.SpinStart("Bootstrapping Agent " + agent.name)
	// Connect to agent over SSH
	if err := agent.ssh.Connect(); err != nil {
		return err
	}
	defer agent.ssh.Disconnect()

	// Copy installation scripts to remote hosts
	installAgentScript := util.GetStaticFile("install_agent.sh")
	reader := strings.NewReader(installAgentScript)
	if err := agent.ssh.CopyTo(reader, "/tmp/", "install_agent.sh", "0775", len(installAgentScript)); err != nil {
		return err
	}

	waitAgentScript := util.GetStaticFile("wait_agent.sh")
	reader = strings.NewReader(waitAgentScript)
	if err := agent.ssh.CopyTo(reader, "/tmp/", "wait_agent.sh", "0775", len(waitAgentScript)); err != nil {
		return err
	}

	// Instantiate install arguments
	installArgs := ""
	pkgCloudToken, pkgExists := os.LookupEnv("PACKAGE_CLOUD_TOKEN")
	agentVersion, verExists := os.LookupEnv("IOFOG_AGENT_VERSION")
	if pkgExists && verExists {
		installArgs = "dev " + agentVersion + " " + pkgCloudToken
	}

	// Define commands
	cmds := []string{
		"echo 'APT::Get::AllowUnauthenticated \"true\";' | sudo -S tee /etc/apt/apt.conf.d/99temp",
		"sudo -S apt --assume-yes install apt-transport-https ca-certificates curl software-properties-common jq",
		"/tmp/install_agent.sh " + installArgs,
		"sudo -S service iofog-agent start",
		"/tmp/wait_agent.sh",
		"sudo -S iofog-agent config -cf 10 -sf 10",
	}

	// Execute commands
	for _, cmd := range cmds {
		_, err := agent.ssh.Run(cmd)
		if err != nil {
			return err
		}
	}

	return nil
}

func (agent *RemoteAgent) Configure(ctrl *config.Controller, user client.User) (uuid string, err error) {
	defer util.SpinStop()
	util.SpinStart("Configuring Agent " + agent.name)

	controllerEndpoint := ctrl.Endpoint

	key, uuid, err := agent.getProvisionKey(controllerEndpoint, user)
	if err != nil {
		return
	}

	// Establish SSH to agent
	err = agent.ssh.Connect()
	if err != nil {
		return
	}

	// Prepare progress bar
	defer agent.ssh.Disconnect()

	// Instantiate commands
	controllerBaseURL := fmt.Sprintf("http://%s/api/v3", controllerEndpoint)
	cmds := []string{
		"sudo iofog-agent config -a " + controllerBaseURL,
		"sudo iofog-agent provision " + key,
	}

	// Execute commands
	for _, cmd := range cmds {
		_, err = agent.ssh.Run(cmd)
		if err != nil {
			return
		}
	}

	return
}
