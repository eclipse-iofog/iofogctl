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
	"fmt"

	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/v3/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

type agentExecutor struct {
	namespace string
	name      string
}

func newAgentExecutor(namespace, name string) *agentExecutor {
	exe := &agentExecutor{}
	exe.namespace = namespace
	exe.name = name
	return exe
}

func (exe *agentExecutor) GetName() string {
	return exe.name
}

func (exe *agentExecutor) Execute() error {
	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	// Update local cache based on Controller
	if err := clientutil.SyncAgentInfo(exe.namespace); err != nil {
		return err
	}

	// Get agent config
	baseAgent, err := ns.GetAgent(exe.name)
	if err != nil {
		return err
	}

	switch agent := baseAgent.(type) {
	case *rsc.LocalAgent:
		lc, err := install.NewLocalContainerClient()
		if err != nil {
			return err
		}
		containerName := install.GetLocalContainerName("agent", false)
		stdout, stderr, err := lc.GetLogsByName(containerName)
		if err != nil {
			return err
		}

		printContainerLogs(stdout, stderr)

		return nil
	case *rsc.RemoteAgent:
		// Establish SSH connection
		if err := agent.ValidateSSH(); err != nil {
			return err
		}
		ssh, err := util.NewSecureShellClient(agent.SSH.User, agent.Host, agent.SSH.KeyFile)
		if err != nil {
			return err
		}
		ssh.SetPort(agent.SSH.Port)
		err = ssh.Connect()
		if err != nil {
			return err
		}

		// Get logs
		out, err := ssh.Run("sudo cat /var/log/iofog-agent/iofog-agent.0.log")
		if err != nil {
			return err
		}
		fmt.Print(out.String())
	}

	return nil
}
