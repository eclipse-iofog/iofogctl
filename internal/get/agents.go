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

	"github.com/eclipse-iofog/iofog-go-sdk/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type agentExecutor struct {
	namespace    string
	showDetached bool
}

func newAgentExecutor(namespace string, showDetached bool) *agentExecutor {
	a := &agentExecutor{}
	a.namespace = namespace
	a.showDetached = showDetached
	return a
}

func (exe *agentExecutor) GetName() string {
	return ""
}

func (exe *agentExecutor) Execute() error {
	if exe.showDetached {
		printDetached()
		if err := generateDetachedAgentOutput(); err != nil {
			return err
		}
		return nil
	}
	printNamespace(exe.namespace)
	if err := generateAgentOutput(exe.namespace); err != nil {
		return err
	}
	return config.Flush()
}

func generateDetachedAgentOutput() error {
	detachedAgents, err := config.GetDetachedAgents()
	if err != nil {
		return err
	}
	// Make an index of agents the client knows about and pre-process any info
	agentsToPrint := make(map[string]client.AgentInfo)
	for _, agent := range detachedAgents {
		agentsToPrint[agent.Name] = client.AgentInfo{
			Name:              agent.Name,
			IPAddressExternal: agent.Host,
		}
	}
	return tabulateAgents(agentsToPrint)
}

func generateAgentOutput(namespace string) error {
	// Get Config
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return err
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
	// Instantiate client
	// Log into Controller
	ctrl, err := internal.NewControllerClient(namespace)
	if err != nil {
		return tabulateAgents(agentsToPrint)
	}

	// Get Agents from Controller
	listAgentsResponse, err := ctrl.ListAgents()
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

	return tabulateAgents(agentsToPrint)
}

func tabulateAgents(agentInfos map[string]client.AgentInfo) error {
	// Generate table and headers
	table := make([][]string, len(agentInfos)+1)
	headers := []string{
		"AGENT",
		"STATUS",
		"AGE",
		"UPTIME",
		"ADDR",
		"VERSION",
	}
	table[0] = append(table[0], headers...)
	// Populate rows
	idx := 0
	for _, agent := range agentInfos {
		// if UUID is empty, we assume the agent is not provisioned
		if agent.UUID == "" {
			row := []string{
				agent.Name,
				"not provisioned",
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
