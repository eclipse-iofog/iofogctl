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

package deleteagent

import (
	"fmt"
	"strings"

	"github.com/eclipse-iofog/iofogctl/internal"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type remoteExecutor struct {
	namespace string
	name      string
}

func newRemoteExecutor(namespace, name string) *remoteExecutor {
	exe := &remoteExecutor{}
	exe.namespace = namespace
	exe.name = name
	return exe
}

func (exe *remoteExecutor) GetName() string {
	return exe.name
}

func (exe *remoteExecutor) Execute() error {
	// Check the agent exists
	agent, err := config.GetAgent(exe.namespace, exe.name)
	if err != nil {
		return err
	}

	ctrl, err := internal.NewControllerClient(exe.namespace)
	// If controller exists, deprovision the agent
	if err == nil {
		// Perform deletion of Agent through Controller
		if err = ctrl.DeleteAgent(agent.UUID); err != nil {
			if !strings.Contains(err.Error(), "NotFoundError") {
				return err
			}
		}
	}

	// Stop the Agent process on remote server
	if agent.Host == "" || agent.SSH.User == "" || agent.SSH.KeyFile == "" || agent.SSH.Port == 0 {
		util.PrintNotify("Could not stop daemon for Agent " + agent.Name + ". SSH details missing from local cofiguration. Use configure command to add SSH details.")
	} else {
		sshAgent := install.NewRemoteAgent(agent.SSH.User, agent.Host, agent.SSH.Port, agent.SSH.KeyFile, agent.Name)
		if err = sshAgent.Stop(); err != nil {
			util.PrintNotify(fmt.Sprintf("Failed to stop daemon on Agent %s. %s", agent.Name, err.Error()))
		}
	}

	// Update config
	if err = config.DeleteAgent(exe.namespace, exe.name); err != nil {
		return err
	}

	return nil
}
