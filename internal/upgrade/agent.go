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

package upgrade

import (
	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	clientutil "github.com/eclipse-iofog/iofogctl/v3/internal/util/client"
)

type agentExecutor struct {
	namespace string
	name      string
}

func newAgentExecutor(opt Options) *agentExecutor {
	return &agentExecutor{
		namespace: opt.Namespace,
		name:      opt.Name,
	}
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

	// Get the Agent to verify it exists
	agent, err := ns.GetAgent(exe.name)
	if err != nil {
		return err
	}

	// Talk to Controller
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	// Perform upgrade
	if err := clt.UpgradeAgent(agent.GetName()); err != nil {
		return err
	}

	return nil
}
