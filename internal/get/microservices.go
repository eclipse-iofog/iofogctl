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
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type microserviceExecutor struct {
	namespace  string
	client     *client.Client
	flows      []client.FlowInfo
	msvcPerID  map[string]*client.MicroserviceInfo
	agentPerID map[string]*client.AgentInfo
}

func newMicroserviceExecutor(namespace string) *microserviceExecutor {
	a := &microserviceExecutor{}
	a.namespace = namespace
	a.msvcPerID = make(map[string]*client.MicroserviceInfo)
	a.agentPerID = make(map[string]*client.AgentInfo)
	return a
}

func (exe *microserviceExecutor) init(controlPlane config.ControlPlane) (err error) {
	// TODO: replace controllers[0] with controplane variable
	exe.client = client.New(controlPlane.Controllers[0].Endpoint)
	if err = exe.client.Login(client.LoginRequest{Email: controlPlane.IofogUser.Email, Password: controlPlane.IofogUser.Password}); err != nil {
		return
	}
	listMsvcs, err := exe.client.GetAllMicroservices()
	if err != nil {
		return err
	}
	for i := 0; i < len(listMsvcs.Microservices); i++ {
		exe.msvcPerID[listMsvcs.Microservices[i].UUID] = &listMsvcs.Microservices[i]
	}

	listAgents, err := exe.client.ListAgents()
	if err != nil {
		return err
	}
	for i := 0; i < len(listAgents.Agents); i++ {
		exe.agentPerID[listAgents.Agents[i].UUID] = &listAgents.Agents[i]
	}
	return
}

func (exe *microserviceExecutor) GetName() string {
	return ""
}

func (exe *microserviceExecutor) Execute() error {
	// Get Control Plane config details
	controlPlane, err := config.GetControlPlane(exe.namespace)
	if err != nil {
		return err
	}
	if len(controlPlane.Controllers) == 0 {
		// Generate empty output
		return exe.generateMicroserviceOutput()
	}
	// Fetch data
	if err = exe.init(controlPlane); err != nil {
		return err
	}

	return exe.generateMicroserviceOutput()
}

func (exe *microserviceExecutor) generateMicroserviceOutput() (err error) {

	// Generate table and headers
	table := make([][]string, len(exe.msvcPerID)+1)
	headers := []string{"MICROSERVICE", "STATUS", "AGENT", "CONFIG", "ROUTES", "VOLUMES", "PORTS"}
	table[0] = append(table[0], headers...)

	// Populate rows
	count := 0
	for _, ms := range exe.msvcPerID {
		if util.IsSystemMsvc(*ms) {
			continue
		}

		routes := ""
		for idx, route := range ms.Routes {
			routeDestName := "unknown"
			routeDest, ok := exe.msvcPerID[route]
			if ok == true {
				routeDestName = routeDest.Name
			}
			if idx == 0 {
				routes += routeDestName
			} else {
				routes += fmt.Sprintf(", %s", routeDestName)
			}
		}
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
				ports += fmt.Sprintf("%d:%d", port.External, port.Internal)
			} else {
				ports += fmt.Sprintf(", %d:%d", port.External, port.Internal)
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
		if status == "" {
			status = "Not Supported"
		}
		row := []string{
			ms.Name,
			status,
			agentName,
			ms.Config,
			routes,
			volumes,
			ports,
		}
		table[count+1] = append(table[count+1], row...)
		count++
	}

	// Print the table
	err = print(table)
	if err != nil {
		return err
	}

	return nil
}
