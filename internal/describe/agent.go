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
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"strings"
)

type agentExecutor struct {
	namespace string
	name      string
}

func newAgentExecutor(namespace, name string) *agentExecutor {
	a := &agentExecutor{}
	a.namespace = namespace
	a.name = name
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
	ctrl := iofog.NewController(ctrls[0].Endpoint)
	loginRequest := iofog.LoginRequest{
		Email:    ctrls[0].IofogUser.Email,
		Password: ctrls[0].IofogUser.Password,
	}

	// Send requests to controller
	loginResponse, err := ctrl.Login(loginRequest)
	if err != nil {
		return err
	}
	token := loginResponse.AccessToken
	getAgentResponse, err := ctrl.GetAgent(agent.UUID, token)
	if err != nil {
		// The agents might not be provisioned with Controller
		if strings.Contains(err.Error(), "NotFoundError") {
			return util.NewInputError("Cannot describe an Agent that is not provisioned with the Controller in namespace " + exe.namespace)
		}
		return err
	}

	// Print result
	if err = util.Print(getAgentResponse); err != nil {
		return err
	}
	return nil
}
