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
	"fmt"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
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

func (exe *agentExecutor) Execute() error {
	// Get agent config
	agent, err := config.GetAgent(exe.namespace, exe.name)
	if err != nil {
		return err
	}

	// Establish SSH connection
	if agent.Host == "" || agent.User == "" || agent.KeyFile == "" || agent.Port == 0 {
		util.Check(util.NewError("Cannot get logs because SSH details for this Agent are not available"))
	}
	ssh := util.NewSecureShellClient(agent.User, agent.Host, agent.KeyFile)
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

	return nil
}
