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

	"github.com/eclipse-iofog/iofogctl/pkg/util"
	pb "github.com/schollz/progressbar"
)

type Agent interface {
	Bootstrap() error
	getProvisionKey(string, User, *pb.ProgressBar) (string, string, error)
	Configure(string, User) (string, error)
}

// defaultAgent implements commong behavior
type defaultAgent struct {
	name string
}

func (agent *defaultAgent) getProvisionKey(controllerEndpoint string, user User, pb *pb.ProgressBar) (key string, uuid string, err error) {
	// Connect to controller
	ctrl := NewController(controllerEndpoint)

	// Log in
	loginRequest := LoginRequest{
		Email:    user.Email,
		Password: user.Password,
	}
	loginResponse, err := ctrl.Login(loginRequest)
	if err != nil {
		return
	}
	token := loginResponse.AccessToken
	pb.Add(20)

	// Create agent
	createRequest := CreateAgentRequest{
		Name:    agent.name,
		FogType: 0,
	}
	createResponse, err := ctrl.CreateAgent(createRequest, token)
	if err != nil {
		return
	}
	uuid = createResponse.UUID
	pb.Add(20)

	// Get provisioning key
	provisionResponse, err := ctrl.GetAgentProvisionKey(uuid, token)
	if err != nil {
		return
	}
	pb.Add(20)
	key = provisionResponse.Key
	return
}

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
		{"echo 'APT::Get::AllowUnauthenticated \"true\";' | sudo tee /etc/apt/apt.conf.d/99temp", 1},
		{"sudo apt --assume-yes install apt-transport-https ca-certificates curl software-properties-common jq", 5},
		{"curl " + installURL + " | sudo tee /opt/linux.sh", 2},
		{"sudo chmod +x /opt/linux.sh", 1},
		{"sudo /opt/linux.sh " + installArgs, 70},
		{"sudo service iofog-agent start", 3},
		{"echo '" + waitForAgentScript + "' | tee ~/wait-for-agent.sh", 1},
		{"sudo chmod +x ~/wait-for-agent.sh", 1},
		{"~/wait-for-agent.sh", 15},
		{"sudo iofog-agent config -cf 10 -sf 10", 1},
	}

	// Prepare progress bar
	pb := pb.New(100)
	defer pb.Clear()

	// Execute commands
	for _, cmd := range cmds {
		_, err = agent.ssh.Run(cmd.cmd)
		pb.Add(cmd.pbSlice)
		if err != nil {
			return err
		}
	}

	return nil
}

func (agent *RemoteAgent) Configure(controllerEndpoint string, user User) (uuid string, err error) {
	pb := pb.New(100)
	defer pb.Clear()

	key, uuid, err := agent.getProvisionKey(controllerEndpoint, user, pb)
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
	pb.Add(20)

	// Instantiate commands
	controllerBaseURL := fmt.Sprintf("http://%s/api/v3", controllerEndpoint)
	cmds := []command{
		{"sudo iofog-agent config -a " + controllerBaseURL, 10},
		{"sudo iofog-agent provision " + key, 10},
	}

	// Execute commands
	for _, cmd := range cmds {
		_, err = agent.ssh.Run(cmd.cmd)
		pb.Add(cmd.pbSlice)
		if err != nil {
			return
		}
	}

	return
}

type command struct {
	cmd     string
	pbSlice int
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
