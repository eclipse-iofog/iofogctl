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

package agent

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	clientutil "github.com/eclipse-iofog/iofogctl/v3/internal/util/client"

	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

func Execute(namespace, name, newName string, useDetached bool) error {
	if err := util.IsLowerAlphanumeric("Agent", newName); err != nil {
		return err
	}
	util.SpinStart(fmt.Sprintf("Renaming Agent %s", name))

	if useDetached {
		if err := config.RenameDetachedAgent(name, newName); err != nil {
			return err
		}
		return config.Flush()
	}

	// Get config
	// Update local cache based on Controller
	if err := clientutil.SyncAgentInfo(namespace); err != nil {
		return err
	}
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return err
	}
	agent, err := ns.GetAgent(name)
	if err != nil {
		return err
	}

	// Init remote resources
	clt, err := clientutil.NewControllerClient(namespace)
	if err != nil {
		return err
	}

	if _, err = clt.UpdateAgent(&client.AgentUpdateRequest{
		UUID: agent.GetUUID(),
		Name: newName,
	}); err != nil {
		return err
	}
	if err := ns.DeleteAgent(name); err != nil {
		return err
	}
	agent.SetName(newName)
	if err := ns.AddAgent(agent); err != nil {
		return err
	}

	return config.Flush()
}
