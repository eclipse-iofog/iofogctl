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

package agent

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func Execute(namespace, name, newName string) error {
	// Get config
	agent, err := config.GetAgent(namespace, name)
	if err != nil {
		return err
	}

	util.SpinStart(fmt.Sprintf("Renaming Agent %s", name))

	// Init remote resources
	clt, err := internal.NewControllerClient(namespace)
	if err != nil {
		return err
	}

	ag, err := config.GetAgent(namespace, name)
	if err != nil {
		return err
	}
	config.DeleteAgent(namespace, name)
	ag.Name = newName
	config.AddAgent(namespace, ag)

	_, err = clt.UpdateAgent(&client.AgentUpdateRequest{
		UUID: agent.UUID,
		Name: newName,
	})

	if err != nil {
		return err
	}
	config.Flush()

	return err
}
