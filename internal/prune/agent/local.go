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

	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func (exe executor) localAgentPrune() error {

	containerClient, err := install.NewLocalContainerClient()
	if err != nil {
		util.PrintNotify(fmt.Sprintf("Could not prune local agent. Error: %s\n", err.Error()))
	} else {
		if _, err = containerClient.ExecuteCmd(install.GetLocalContainerName("agent"), []string{
			"sudo",
			"iofog-agent",
			"prune",
		}); err != nil {
			util.PrintNotify(fmt.Sprintf("Could not prune local agent. Error: %s\n", err.Error()))
		}
	}
	return nil
}