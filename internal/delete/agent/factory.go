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
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func NewExecutor(namespace, name string) (execute.Executor, error) {
	// Check the agent exists
	agent, err := config.GetAgent(namespace, name)
	if err != nil {
		return nil, err
	}

	// Local executor
	if util.IsLocalHost(agent.Host) {
		cli, err := install.NewLocalContainerClient()
		if err != nil {
			return nil, err
		}
		return newLocalExecutor(namespace, name, cli), nil
	}

	// Default executor
	if agent.Host == "" || agent.KeyFile == "" || agent.User == "" {
		return nil, util.NewError("Cannot execute delete command because SSH details for Agent " + name + " are not available")
	}
	return newRemoteExecutor(namespace, name), nil
}
