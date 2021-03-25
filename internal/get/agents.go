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
	"fmt"
	"time"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/v3/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
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
		table, err := generateDetachedAgentOutput()
		if err != nil {
			return err
		}
		return print(table)
	}
	printNamespace(exe.namespace)
	table, err := generateAgentOutput(exe.namespace)
	if err != nil {
		return err
	}
	// Flush occurs in generateAgentOutput
	return print(table)
}

func generateDetachedAgentOutput() (table [][]string, err error) {
	detachedAgents := config.GetDetachedAgents()
	// Make an index of agents the client knows about and pre-process any info
	agentsToPrint := make([]client.AgentInfo, len(detachedAgents))
	for idx := range detachedAgents {
		agentsToPrint[idx] = client.AgentInfo{
			Name:              detachedAgents[idx].GetName(),
			IPAddressExternal: detachedAgents[idx].GetHost(),
		}
	}
	return tabulateAgents(agentsToPrint)
}

func generateAgentOutput(namespace string) (table [][]string, err error) {
	agents := []client.AgentInfo{}
	// Update local cache based on Controller
	if err = clientutil.SyncAgentInfo(namespace); err != nil && !rsc.IsNoControlPlaneError(err) {
		return
	}

	// Get Agents from Controller
	if err == nil {
		agents, err = clientutil.GetBackendAgents(namespace)
		if err != nil {
			return
		}
	}

	return tabulateAgents(agents)
}

func tabulateAgents(agentInfos []client.AgentInfo) (table [][]string, err error) {
	// Generate table and headers
	table = make([][]string, len(agentInfos)+1)
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
	for idx := range agentInfos {
		agent := &agentInfos[idx]
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
			age := "-"
			if backendAge, err := util.ElapsedRFC(agent.CreatedTimeRFC3339, util.NowRFC()); err == nil {
				age = backendAge
			}
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
	}
	return table, err
}

func printDetached() {
	fmt.Printf("DETACHED RESOURCES\n\n")
}
