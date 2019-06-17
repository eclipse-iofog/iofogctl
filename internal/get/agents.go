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

package get

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"strings"
	"time"
)

type agentExecutor struct {
	namespace string
}

func newAgentExecutor(namespace string) *agentExecutor {
	a := &agentExecutor{}
	a.namespace = namespace
	return a
}

func (exe *agentExecutor) Execute() error {
	// Get Config
	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	if len(ns.Controllers) > 1 {
		return util.NewInternalError("Expected 0 or 1 controller in namespace " + exe.namespace)
	}

	// Generate table and headers
	table := make([][]string, len(ns.Agents)+1)
	headers := []string{
		"AGENT",
		"STATUS",
		"AGE",
		"UPTIME",
		"IP",
	}
	table[0] = append(table[0], headers...)

	// Connect to controller if it is ready
	agentInfos := make([]iofog.AgentInfo, len(ns.Agents))
	if len(ns.Controllers) > 0 && ns.Controllers[0].Endpoint != "" {
		ctrl := iofog.NewController(ns.Controllers[0].Endpoint)
		loginRequest := iofog.LoginRequest{
			Email:    ns.Controllers[0].IofogUser.Email,
			Password: ns.Controllers[0].IofogUser.Password,
		}
		// Send requests to controller
		loginResponse, err := ctrl.Login(loginRequest)
		if err != nil {
			return err
		}
		token := loginResponse.AccessToken

		// Get agents from Controller
		for idx, agent := range ns.Agents {
			agentInfo, err := ctrl.GetAgent(agent.UUID, token)
			if err != nil {
				// The agents might not be provisioned with Controller
				if strings.Contains(err.Error(), "NotFoundError") {
					continue
				}
				return err
			}
			agentInfos[idx] = agentInfo
		}
	}

	// Populate rows
	for idx, agent := range ns.Agents {
		age, err := util.ElapsedRFC(agentInfos[idx].CreatedTimeRFC3339, util.NowRFC())
		if err != nil {
			return err
		}
		uptime := time.Duration(agentInfos[idx].DaemonUptimeDurationMsUTC)
		row := []string{
			agent.Name,
			agentInfos[idx].DaemonStatus,
			age,
			util.FormatDuration(uptime),
			agentInfos[idx].IPAddress,
		}
		table[idx+1] = append(table[idx+1], row...)
	}

	// Print table
	err = print(table)
	if err != nil {
		return err
	}

	return nil
}
