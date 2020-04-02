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

package deleteagent

import (
	"fmt"

	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

func (exe executor) deleteRemoteAgent(agent *rsc.RemoteAgent) (err error) {
	// Stop and remove the Agent process on remote server
	if agent.Host == "" || agent.SSH.User == "" || agent.SSH.KeyFile == "" || agent.SSH.Port == 0 {
		util.PrintNotify("Could not stop daemon for Agent " + agent.Name + ". SSH details missing from local cofiguration. Use configure command to add SSH details.")
	} else {
		sshAgent := install.NewRemoteAgent(
			agent.SSH.User,
			agent.Host,
			agent.SSH.Port,
			agent.SSH.KeyFile,
			agent.Name,
			agent.UUID)
		if err = sshAgent.Uninstall(); err != nil {
			util.PrintNotify(fmt.Sprintf("Failed to stop daemon on Agent %s. %s", agent.Name, err.Error()))
		}
	}
	return
}
