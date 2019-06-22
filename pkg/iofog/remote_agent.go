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
	"os"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	pb "github.com/schollz/progressbar"
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
		defaultAgent: defaultAgent{name: agentName, pb: pb.New(100)},
		ssh:          ssh,
	}
}

func (agent *RemoteAgent) Bootstrap() error {
	// Connect to agent over SSH
	err := agent.ssh.Connect()
	if err != nil {
		return err
	}
	defer agent.ssh.Disconnect()

	// Instantiate install arguments
	installURL := "https://raw.githubusercontent.com/eclipse-iofog/platform/develop/infrastructure/ansible/scripts/agent.sh"
	installArgs := ""
	pkgCloudToken, pkgExists := os.LookupEnv("PACKAGE_CLOUD_TOKEN")
	agentVersion, verExists := os.LookupEnv("AGENT_VERSION")
	if pkgExists && verExists {
		installArgs = "dev " + agentVersion + " " + pkgCloudToken
	}

	// Execute commands
	cmds := []command{
		{"echo 'APT::Get::AllowUnauthenticated \"true\";' | sudo -S tee /etc/apt/apt.conf.d/99temp", 1},
		{"sudo -S apt --assume-yes install apt-transport-https ca-certificates curl software-properties-common jq", 5},
		{"curl " + installURL + " | sudo  -S tee /opt/linux.sh", 2},
		{"sudo -S chmod +x /opt/linux.sh", 1},
		{"sudo -S /opt/linux.sh " + installArgs, 70},
		{"sudo -S service iofog-agent start", 3},
		{"echo '" + waitForAgentScript + "' > ~/wait-for-agent.sh", 1},
		{"chmod +x ~/wait-for-agent.sh", 1},
		{"~/wait-for-agent.sh", 15},
		{"sudo -S iofog-agent config -cf 10 -sf 10", 1},
	}

	// Prepare progress bar
	pb := pb.New(100)
	defer pb.Clear()

	// Execute commands
	for _, cmd := range cmds {
		util.PrintInfo(cmd.cmd)
		_, err = agent.ssh.Run(cmd.cmd)
		pb.Add(cmd.pbSlice)
		if err != nil {
			util.PrintError(err.Error())
			return err
		}
	}

	return nil
}

func (agent *RemoteAgent) Configure(ctrl *config.Controller, user User) (uuid string, err error) {
	agent.pb = pb.New(100)
	defer agent.pb.Clear()

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
	agent.pb.Add(20)

	// Instantiate commands
	controllerBaseURL := fmt.Sprintf("http://%s/api/v3", controllerEndpoint)
	cmds := []command{
		{"sudo iofog-agent config -a " + controllerBaseURL, 10},
		{"sudo iofog-agent provision " + key, 10},
	}

	// Execute commands
	for _, cmd := range cmds {
		_, err = agent.ssh.Run(cmd.cmd)
		agent.pb.Add(cmd.pbSlice)
		if err != nil {
			return
		}
	}

	return
}

var waitForAgentScript = `STATUS=""
ITER=0
while [ "$STATUS" != "RUNNING" ] ; do
    ITER=$((ITER+1))
    if [ "$ITER" -gt 30 ]; then
        echo 'Timed out waiting for Agent to be RUNNING'
        exit 1
    fi
    sleep 1
    STATUS=$(sudo iofog-agent status | cut -f2 -d: | head -n 1 | tr -d '[:space:]')
done
exit 0`
