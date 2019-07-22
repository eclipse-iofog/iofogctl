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
	"time"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/client"
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

	// Make an index of agents the client knows about and pre-process any info
	agentsToPrint := make(map[string]client.AgentInfo)
	for _, agent := range ns.Agents {
		agentsToPrint[agent.Name] = client.AgentInfo{
			Name:              agent.Name,
			IPAddressExternal: agent.Host,
		}
	}

	// Connect to Controller if it is ready
	if len(ns.Controllers) > 0 && ns.Controllers[0].Endpoint != "" {
		// Instantiate client
		ctrl := client.New(ns.Controllers[0].Endpoint)
		loginRequest := client.LoginRequest{
			Email:    ns.Controllers[0].IofogUser.Email,
			Password: ns.Controllers[0].IofogUser.Password,
		}
		// Log into Controller
		loginResponse, err := ctrl.Login(loginRequest)
		if err != nil {
			return tabulate(agentsToPrint)
		}
		token := loginResponse.AccessToken

		// Get Agents from Controller
		listAgentsResponse, err := ctrl.ListAgents(token)
		if err != nil {
			return err
		}

		// Process Agents
		for _, remoteAgent := range listAgentsResponse.Agents {
			// Server may have agents that the client is not aware of, update config if so
			if _, exists := agentsToPrint[remoteAgent.Name]; !exists {
				newAgentConf := config.Agent{
					Name: remoteAgent.Name,
					UUID: remoteAgent.UUID,
					Host: remoteAgent.IPAddressExternal,
				}
				config.AddAgent(namespace, newAgentConf)
			}

			// Use the pre-processed default info if necessary
			if remoteAgent.IPAddressExternal == "0.0.0.0" {
				remoteAgent.IPAddressExternal = agentsToPrint[remoteAgent.Name].IPAddressExternal
			}

			// Add details for output
			agentsToPrint[remoteAgent.Name] = remoteAgent
		}
	}

	return tabulate(agentsToPrint)
}

func tabulate(agentInfos map[string]client.AgentInfo) error {
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
	idx := 0
	for _, agent := range agentInfos {
		// if UUID is empty, we assume the agent is not provided
		if agent.UUID == "" {
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
		idx = idx + 1
	}

	// Print table
	return print(table)
}
