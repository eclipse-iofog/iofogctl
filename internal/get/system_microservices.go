/*
 *  *******************************************************************************
 *  * Copyright (c) 2023 Contributors to the Eclipse ioFog Project
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
	"math"
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
)

type systemMicroserviceExecutor struct {
	namespace  string
	client     *client.Client
	msvcPerID  map[string]*client.MicroserviceInfo
	agentPerID map[string]*client.AgentInfo
}

func newSystemMicroserviceExecutor(namespace string) *systemMicroserviceExecutor {
	a := &systemMicroserviceExecutor{}
	a.namespace = namespace
	a.msvcPerID = make(map[string]*client.MicroserviceInfo)
	a.agentPerID = make(map[string]*client.AgentInfo)
	return a
}

func (exe *systemMicroserviceExecutor) init() (err error) {
	exe.client, err = clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		if rsc.IsNoControlPlaneError(err) {
			return nil
		}
		return
	}
	listMsvcs, err := exe.client.GetAllSystemMicroservices()
	if err != nil {
		return err
	}
	for i := 0; i < len(listMsvcs.Microservices); i++ {
		exe.msvcPerID[listMsvcs.Microservices[i].UUID] = &listMsvcs.Microservices[i]
	}

	listAgents, err := exe.client.ListAgents(client.ListAgentsRequest{})
	if err != nil {
		return err
	}
	for i := 0; i < len(listAgents.Agents); i++ {
		exe.agentPerID[listAgents.Agents[i].UUID] = &listAgents.Agents[i]
	}
	return
}

func (exe *systemMicroserviceExecutor) GetName() string {
	return ""
}

func (exe *systemMicroserviceExecutor) Execute() error {
	// Fetch data
	if err := exe.init(); err != nil {
		return err
	}
	printNamespace(exe.namespace)
	table := exe.generateMicroserviceOutput()
	return print(table)
}

func (exe *systemMicroserviceExecutor) generateMicroserviceOutput() (table [][]string) {
	// Generate table and headers
	table = make([][]string, len(exe.msvcPerID)+1)
	headers := []string{"SYS-MICROSERVICE", "STATUS", "AGENT", "VOLUMES", "PORTS"}
	table[0] = append(table[0], headers...)

	// Populate rows
	count := 0
	for _, ms := range exe.msvcPerID {
		// if util.IsSystemMsvc(ms) {
		// 	continue
		// }

		volumes := ""
		for idx, volume := range ms.Volumes {
			if idx == 0 {
				volumes += fmt.Sprintf("%s:%s", volume.HostDestination, volume.ContainerDestination)
			} else {
				volumes += fmt.Sprintf(", %s:%s", volume.HostDestination, volume.ContainerDestination)
			}
		}
		ports := ""
		for idx, port := range ms.Ports {
			if idx == 0 {
				ports += fmt.Sprintf("%v:%v", port.External, port.Internal)
			} else {
				ports += fmt.Sprintf(", %v:%v", port.External, port.Internal)
			}
		}
		agent, ok := exe.agentPerID[ms.AgentUUID]
		var agentName string
		if !ok {
			agentName = "-"
		} else {
			agentName = agent.Name
		}
		status := ms.Status.Status
		switch status {
		case "":
			status = "-"
		case "PULLING":
			if ms.Status.Percentage > 0 {
				status = fmt.Sprintf("%s (%d%s)", ms.Status.Status, int(math.Round(ms.Status.Percentage)), "%")
			}
		}
		if ms.Status.ErrorMessage != "" {
			msg := ms.Status.ErrorMessage
			if strings.Contains(msg, "invalid mount config for type \"bind\"") {
				msg = "Volume missing"
			} else if strings.Contains(msg, "runtime create failed") {
				msg = "Error starting container"
			}
			status = fmt.Sprintf("%s (%s)", ms.Status.Status, msg)
		}

		row := []string{
			ms.Name,
			status,
			agentName,
			volumes,
			ports,
		}
		table[count+1] = append(table[count+1], row...)
		count++
	}

	return table
}
