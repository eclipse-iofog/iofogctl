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

package pruneagent

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/v3/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

func (exe executor) localAgentPrune() error {
	containerClient, err := install.NewLocalContainerClient()
	if err != nil {
		return err
	}
	if _, err = containerClient.ExecuteCmd(install.GetLocalContainerName("agent", false), []string{
		"sudo",
		"iofog-agent",
		"prune",
	}); err != nil {
		return util.NewInternalError(fmt.Sprintf("Could not prune local agent. Error: %s\n", err.Error()))
	}

	return nil
}
