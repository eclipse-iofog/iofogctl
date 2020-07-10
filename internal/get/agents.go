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

package get

import (
	"time"

	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	iutil "github.com/eclipse-iofog/iofogctl/v2/internal/util"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
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
	if err := generateAgentOutput(exe.namespace, true); err != nil {
		return err
	}
	// Flush occurs in generateAgentOutput
	return nil
}

func generateDetachedAgentOutput() error {
	detachedAgents := config.GetDetachedAgents()
	// Make an index of agents the client knows about and pre-process any info
	agentsToPrint := make([]client.AgentInfo, 0)
	for _, agent := range detachedAgents {
		agentsToPrint = append(agentsToPrint, client.AgentInfo{
			Name:              agent.GetName(),
			IPAddressExternal: agent.GetHost(),
		})
	}
	return tabulateAgents(agentsToPrint)
}

func generateAgentOutput(namespace string, printNS bool) error {
	agents := make([]client.AgentInfo, 0)
	// Update local cache based on Controller
	err := iutil.UpdateAgentCache(namespace)
	if err != nil && !rsc.IsNoControlPlaneError(err) {
		return err
	}

	// Get Agents from Controller
	if err == nil {
		agents, err = iutil.GetBackendAgents(namespace)
		if err != nil {
			return err
		}
	}

	if printNS {
		printNamespace(namespace)
	}

	return tabulateAgents(agents)
}

func tabulateAgents(agentInfos []client.AgentInfo) error {
	// Generate table and headers
	table := make([][]string, len(agentInfos)+1)
	headers := []string{
		"AGENT",
		"STATUS",
		"AGE",
		"UPTIME",
		"VERSION",
		"ADDR",
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
				"-",
				agent.IPAddressExternal,
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
				agent.Version,
				agent.Host,
			}
			table[idx+1] = append(table[idx+1], row...)
		}
		idx = idx + 1
	}

	// Print table
	return print(table)
}
