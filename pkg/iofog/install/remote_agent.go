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
	"strings"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

// Remote agent uses SSH
type RemoteAgent struct {
	defaultAgent
	ssh     *util.SecureShellClient
	version string
	repo    string
	token   string
}

func NewRemoteAgent(user, host string, port int, privKeyFilename, agentName string) *RemoteAgent {
	ssh := util.NewSecureShellClient(user, host, privKeyFilename)
	ssh.SetPort(port)
	return &RemoteAgent{
		defaultAgent: defaultAgent{name: agentName},
		ssh:          ssh,
		version:      util.GetAgentTag(),
	}
}

func (agent *RemoteAgent) SetVersion(version string) {
	if version == "" {
		return
	}
	agent.version = version
}

func (agent *RemoteAgent) SetRepository(repo, token string) {
	if repo == "" {
		return
	}
	agent.repo = repo
	agent.token = token
}

func (agent *RemoteAgent) Bootstrap() error {
	// Prepare Agent for bootstrap
	if err := agent.copyScriptsToAgent(); err != nil {
		return err
	}

	// Define bootstrap commands
	installArgs := agent.version + " " + agent.repo + " " + agent.token
	cmds := []command{
		{
			cmd: "/tmp/agent_install_java.sh ",
			msg: "Installing Java on Agent " + agent.name,
		},
		{
			cmd: "/tmp/agent_install_docker.sh ",
			msg: "Installing Docker on Agent " + agent.name,
		},
		{
			cmd: "sudo -S /tmp/agent_install_iofog.sh " + installArgs,
			msg: "Installing ioFog daemon on Agent " + agent.name,
		},
		{
			cmd: "sudo -S service iofog-agent start",
			msg: "Starting Agent " + agent.name,
		},
		{
			cmd: "/tmp/agent_wait.sh",
			msg: "Waiting for Agent " + agent.name,
		},
		{
			cmd: "sudo -S iofog-agent config -cf 10 -sf 10",
			msg: "Configuring Agent frequencies",
		},
	}

	// Execute commands on remote server
	if err := agent.run(cmds); err != nil {
		return err
	}

	return nil
}

func (agent *RemoteAgent) Configure(controllerEndpoint string, user IofogUser) (uuid string, err error) {
	key, uuid, err := agent.getProvisionKey(controllerEndpoint, user)
	if err != nil {
		return
	}

	// Instantiate commands
	controllerBaseURL := fmt.Sprintf("http://%s/api/v3", controllerEndpoint)
	cmds := []command{
		{
			cmd: "sudo iofog-agent config -a " + controllerBaseURL,
			msg: "Configuring Agent " + agent.name + " with Controller URL " + controllerBaseURL,
		},
		{
			cmd: "sudo iofog-agent provision " + key,
			msg: "Provisioning Agent " + agent.name + " with Controller",
		},
	}

	// Execute commands on remote server
	if err = agent.run(cmds); err != nil {
		return
	}

	return
}

func (agent *RemoteAgent) Stop() (err error) {
	// Prepare commands
	cmds := []command{
		{
			cmd: "sudo -S service iofog-agent stop",
			msg: "Stopping Agent " + agent.name,
		},
	}

	// Execute commands on remote server
	if err = agent.run(cmds); err != nil {
		return
	}

	return
}

func (agent *RemoteAgent) run(cmds []command) (err error) {
	// Establish SSH to agent
	if err = agent.ssh.Connect(); err != nil {
		return
	}
	defer agent.ssh.Disconnect()

	// Execute commands
	for _, cmd := range cmds {
		verbose(cmd.msg)
		if _, err = agent.ssh.Run(cmd.cmd); err != nil {
			return
		}
	}

	return
}

func (agent RemoteAgent) copyScriptsToAgent() error {
	verbose("Copying install scripts to Agent " + agent.name)
	// Establish SSH to agent
	if err := agent.ssh.Connect(); err != nil {
		return err
	}
	defer agent.ssh.Disconnect()

	// Declare scripts to copy
	scripts := []string{
		"agent_init.sh",
		"agent_install_java.sh",
		"agent_install_docker.sh",
		"agent_install_iofog.sh",
		"agent_wait.sh",
	}
	// Copy scripts to remote host
	for _, script := range scripts {
		staticFile := util.GetStaticFile(script)
		reader := strings.NewReader(staticFile)
		if err := agent.ssh.CopyTo(reader, "/tmp/", script, "0775", len(staticFile)); err != nil {
			return err
		}
	}

	return nil
}

type command struct {
	cmd string
	msg string
}
