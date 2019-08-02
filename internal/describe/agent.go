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
	"fmt"
	"strings"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/client"
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

func (exe *agentExecutor) Execute() error {
	// Get config
	agent, err := config.GetAgent(exe.namespace, exe.name)
	if err != nil {
		return err
	}
	ctrls, err := config.GetControllers(exe.namespace)
	if err != nil {
		return err
	}
	if len(ctrls) != 1 {
		return util.NewInputError("Cannot get Agent data without a Controller in namespace " + exe.namespace)
	}

	// Connect to controller
	ctrl := client.New(ctrls[0].Endpoint)
	loginRequest := client.LoginRequest{
		Email:    ctrls[0].IofogUser.Email,
		Password: ctrls[0].IofogUser.Password,
	}

	// Send requests to controller
	if err := ctrl.Login(loginRequest); err != nil {
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

	// Print result
	fmt.Printf("namespace: %s\n", exe.namespace)
	if err = util.Print(getAgentResponse, exe.filename); err != nil {
		return err
	}
	return nil
}
