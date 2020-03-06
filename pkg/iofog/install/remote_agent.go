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
	"net"
	"net/url"
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/pkg/client"
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

func NewRemoteAgent(user, host string, port int, privKeyFilename, agentName, agentUUID string) *RemoteAgent {
	ssh := util.NewSecureShellClient(user, host, privKeyFilename)
	ssh.SetPort(port)
	return &RemoteAgent{
		defaultAgent: defaultAgent{name: agentName, uuid: agentUUID},
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
	if err := agent.copyInstallScriptsToAgent(); err != nil {
		return err
	}

	// Define bootstrap commands
	installArgs := agent.version + " " + agent.repo + " " + agent.token
	cmds := []command{
		{
			cmd: "/tmp/check_prereqs.sh ",
			msg: "Checking prerequisites on Agent " + agent.name,
		},
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

	// Generate controller endpoint
	u, err := url.Parse(controllerEndpoint)
	if err != nil || u.Host == "" {
		u, err = url.Parse("//" + controllerEndpoint) // Try to see if controllerEndpoint is an IP, in which case it needs to be pefixed by //
		if err != nil {
			return "", err
		}
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	_, _, err = net.SplitHostPort(u.Host) // Returns error if port is not specified
	if err != nil {
		u.Host = u.Host + ":" + client.ControllerPortString
	}
	u.Path = "api/v3"
	u.RawQuery = ""
	u.Fragment = ""
	controllerBaseURL := u.String()
	// Instantiate commands
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

func (agent *RemoteAgent) Deprovision() (err error) {
	// Prepare commands
	cmds := []command{
		{
			cmd: "sudo iofog-agent deprovision",
			msg: "Deprovisioning Agent " + agent.name,
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
			cmd: "sudo iofog-agent deprovision",
			msg: "Deprovisioning Agent " + agent.name,
		},
		{
			cmd: "sudo -S service iofog-agent stop",
			msg: "Stopping Agent " + agent.name,
		},
	}

	// Execute commands on remote server
	if err = agent.run(cmds); err != nil {
		return err
	}

	return
}

func (agent *RemoteAgent) Prune() (err error) {
	// Prepare commands
	cmds := []command{
		{
			cmd: "sudo -S service iofog-agent prune",
			msg: "Pruning Agent " + agent.name,
		},
	}

	// Execute commands on remote server
	if err = agent.run(cmds); err != nil {
		return err
	}

	return
}

func (agent *RemoteAgent) Uninstall() (err error) {
	// Prepare Agent for removal
	if err := agent.copyUninstallScriptsToAgent(); err != nil {
		return err
	}
	// Stop iofog-agent properly
	if err = agent.Stop(); err != nil {
		return
	}

	// Prepare commands
	cmds := []command{
		// TODO: Implement purge on agent
		// {
		// 	cmd: "sudo iofog-agent purge",
		// 	msg: "Deprovisioning Agent " + agent.name,
		// },
		{
			cmd: "/tmp/agent_uninstall_iofog.sh ",
			msg: "Removing iofog-agent software " + agent.name,
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
		Verbose(cmd.msg)
		if _, err = agent.ssh.Run(cmd.cmd); err != nil {
			return err
		}
	}

	return
}

func (agent RemoteAgent) copyInstallScriptsToAgent() error {
	Verbose("Copying install scripts to Agent " + agent.name)
	// Declare scripts to copy
	scripts := []string{
		"check_prereqs.sh",
		"agent_init.sh",
		"agent_install_java.sh",
		"agent_install_docker.sh",
		"agent_install_iofog.sh",
		"agent_wait.sh",
	}
	return agent.copyScriptsToAgent(scripts)
}

func (agent RemoteAgent) copyUninstallScriptsToAgent() error {
	Verbose("Copying uninstall scripts to Agent " + agent.name)
	// Declare scripts to copy
	scripts := []string{
		"agent_init.sh",
		"agent_uninstall_iofog.sh",
	}
	return agent.copyScriptsToAgent(scripts)
}

func (agent RemoteAgent) copyScriptsToAgent(scripts []string) error {
	// Establish SSH to agent
	if err := agent.ssh.Connect(); err != nil {
		return err
	}
	defer agent.ssh.Disconnect()

	// Copy scripts to remote host
	for _, script := range scripts {
		staticFile := util.GetStaticFile(script)
		reader := strings.NewReader(staticFile)
		if err := agent.ssh.CopyTo(reader, "/tmp/", script, "0775", int64(len(staticFile))); err != nil {
			return err
		}
	}

	return nil
}

type command struct {
	cmd string
	msg string
}
