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

package pruneagent

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func (exe executor) remoteAgentPrune(agent config.Agent) error {
	if agent.Host == "" || agent.SSH.User == "" || agent.SSH.KeyFile == "" || agent.SSH.Port == 0 {
		util.PrintNotify("Could not prune agent " + agent.Name + ". SSH details missing from local configuration. Use configure command to add SSH details.")
	} else {
		sshAgent := install.NewRemoteAgent(agent.SSH.User, agent.Host, agent.SSH.Port, agent.SSH.KeyFile, agent.Name, nil)
		if err := sshAgent.Prune(); err != nil {
			util.PrintNotify(fmt.Sprintf("Could not prune agent %s. %s", agent.Name, err.Error()))
		}
	}
	return nil
}
