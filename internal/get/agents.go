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
	"strings"
	"time"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
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
	printNamespace(exe.namespace)
	return generateAgentOutput(exe.namespace)
}

func generateAgentOutput(namespace string) error {
	// Get Config
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return err
	}
	if len(ns.Controllers) > 1 {
		return util.NewInternalError("Expected 0 or 1 controller in namespace " + namespace)
	}

	// Pre process the output with agent names
	agentInfos := make([]iofog.AgentInfo, len(ns.Agents))
	for idx, agent := range ns.Agents {
		agentInfos[idx].Name = agent.Name
		agentInfos[idx].IPAddressExternal = agent.Host
	}

	// Connect to controller if it is ready
	if len(ns.Controllers) > 0 && ns.Controllers[0].Endpoint != "" {
		ctrl := iofog.NewController(ns.Controllers[0].Endpoint)
		loginRequest := iofog.LoginRequest{
			Email:    ns.Controllers[0].IofogUser.Email,
			Password: ns.Controllers[0].IofogUser.Password,
		}
		// Send requests to controller
		loginResponse, err := ctrl.Login(loginRequest)
		if err != nil {
			return tabulate(agentInfos)
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
				return tabulate(agentInfos)
			}
			if agentInfo.IPAddressExternal == "0.0.0.0" {
				agentInfo.IPAddressExternal = agent.Host
			}
			agentInfos[idx] = agentInfo
		}
	}

	return tabulate(agentInfos)
}

func tabulate(agentInfos []iofog.AgentInfo) error {
	// Generate table and headers
	table := make([][]string, len(agentInfos)+1)
	headers := []string{
		"AGENT",
		"STATUS",
		"AGE",
		"UPTIME",
		"IP",
		"VERSION",
	}
	table[0] = append(table[0], headers...)
	// Populate rows
	for idx, agent := range agentInfos {
		// if UUID is empty, we assume the agent is not provided
		if agentInfos[idx].UUID == "" {
			row := []string{
				agent.Name,
				"offline",
				"-",
				"-",
				agent.IPAddressExternal,
				"-",
			}
			table[idx+1] = append(table[idx+1], row...)
		} else {
			age, _ := util.ElapsedRFC(agent.CreatedTimeRFC3339, util.NowRFC())
			uptime := time.Duration(agent.UptimeMs) * time.Millisecond
			row := []string{
				agent.Name,
				agent.DaemonStatus,
				age,
				util.FormatDuration(uptime),
				agent.IPAddressExternal,
				agent.Version,
			}
			table[idx+1] = append(table[idx+1], row...)
		}
	}

	// Print table
	return print(table)
}
