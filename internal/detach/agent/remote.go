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

package detachagent

import (
	"fmt"

	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

func (exe executor) remoteDeprovision(agent *rsc.RemoteAgent) error {
	if agent.ValidateSSH() != nil {
		util.PrintNotify("Could not deprovision daemon for Agent " + agent.Name + ". SSH details missing from local configuration. Use configure command to add SSH details.")
	} else {
		sshAgent, err := install.NewRemoteAgent(
			agent.SSH.User,
			agent.Host,
			agent.SSH.Port,
			agent.SSH.KeyFile,
			agent.Name,
			agent.UUID)
		if err != nil {
			return err
		}
		if err := sshAgent.Deprovision(); err != nil {
			util.PrintNotify(fmt.Sprintf("Failed to deprovision daemon on Agent %s. %s", agent.Name, err.Error()))
		}
	}
	return nil
}
