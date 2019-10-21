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

package describe

import (
	"strings"

	"github.com/eclipse-iofog/iofogctl/internal"

	apps "github.com/eclipse-iofog/iofog-go-sdk/pkg/apps"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type agentExecutor struct {
	namespace string
	name      string
	filename  string
}

func newAgentExecutor(namespace, name, filename string) *agentExecutor {
	a := &agentExecutor{}
	a.namespace = namespace
	a.name = name
	a.filename = filename
	return a
}

func (exe *agentExecutor) GetName() string {
	return exe.name
}

func (exe *agentExecutor) Execute() error {
	// Get config
	agent, err := config.GetAgent(exe.namespace, exe.name)
	if err != nil {
		return err
	}

	// Connect to controller
	ctrl, err := internal.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	getAgentResponse, err := ctrl.GetAgentByID(agent.UUID)
	if err != nil {
		// The agents might not be provisioned with Controller
		if strings.Contains(err.Error(), "NotFoundError") {
			return util.NewInputError("Cannot describe an Agent that is not provisioned with the Controller in namespace " + exe.namespace)
		}
		return err
	}

	header := config.Header{
		Kind: apps.AgentKind,
		Metadata: config.HeaderMetadata{
			Namespace: exe.namespace,
			Name:      exe.name,
		},
		Spec: getAgentResponse,
	}

	if exe.filename == "" {
		if err = util.Print(header); err != nil {
			return err
		}
	} else {
		if err = util.FPrint(header, exe.filename); err != nil {
			return err
		}
	}
	return nil
}
