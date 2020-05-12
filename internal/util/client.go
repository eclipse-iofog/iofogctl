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

package util

import (
	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
)

var cachedClient *client.Client
var cachedAgents []client.AgentInfo

// NewControllerClient returns an iofog-go-sdk/client configured for the current namespace
func NewControllerClient(namespace string) (*client.Client, error) {
	if cachedClient != nil {
		return cachedClient, nil
	}
	// Get endpoint
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return nil, err
	}
	controlPlane, err := ns.GetControlPlane()
	if err != nil {
		return nil, err
	}
	endpoint, err := controlPlane.GetEndpoint()
	if err != nil {
		return nil, err
	}

	user := controlPlane.GetUser()
	cachedClient, err = client.NewAndLogin(client.Options{Endpoint: endpoint}, user.Email, user.GetRawPassword())
	if err != nil {
		return nil, err
	}

	return cachedClient, nil
}

func GetBackendAgents(namespace string) ([]client.AgentInfo, error) {
	if cachedAgents != nil {
		return cachedAgents, nil
	}
	ioClient, err := NewControllerClient(namespace)
	if err != nil {
		return nil, err
	}
	agentList, err := ioClient.ListAgents(client.ListAgentsRequest{})
	if err != nil {
		return nil, err
	}
	cachedAgents = agentList.Agents
	return cachedAgents, nil
}

func UpdateAgentCache(namespace string) error {
	// Get local cache Agents
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return err
	}
	// Check the Control Plane type
	controlPlane, err := ns.GetControlPlane()
	if err != nil {
		return err
	}
	switch controlPlane.(type) {
	case *rsc.LocalControlPlane:
		// Do not update local Agents
		return nil
	}
	agentsMap := make(map[string]*rsc.RemoteAgent, 0)
	for _, baseAgent := range ns.GetAgents() {
		agentsMap[baseAgent.GetName()] = baseAgent.(*rsc.RemoteAgent)
	}
	// Get backend Agents
	backendAgents, err := GetBackendAgents(namespace)
	if err != nil {
		return err
	}

	// Generate cache types
	agents := make([]rsc.RemoteAgent, 0)
	for _, backendAgent := range backendAgents {
		agent := rsc.RemoteAgent{
			Name: backendAgent.Name,
			UUID: backendAgent.UUID,
			Host: backendAgent.Host,
		}
		// Update additional info if local cache contains it
		if cachedAgent, exists := agentsMap[backendAgent.Name]; exists {
			agent.Created = cachedAgent.GetCreatedTime()
			agent.SSH = cachedAgent.SSH
		}

		agents = append(agents, agent)
	}

	// Overwrite the Agents
	ns.DeleteAgents()
	for idx := range agents {
		if err := ns.AddAgent(&agents[idx]); err != nil {
			return err
		}
	}

	return config.Flush()
}

func GetMicroserviceName(namespace, uuid string) (name string, err error) {
	clt, err := NewControllerClient(namespace)
	if err != nil {
		return
	}

	response, err := clt.GetMicroserviceByID(uuid)
	if err != nil {
		return
	}

	name = response.Name
	return
}

func GetMicroserviceUUID(namespace, name string) (uuid string, err error) {
	clt, err := NewControllerClient(namespace)
	if err != nil {
		return
	}

	response, err := clt.GetMicroserviceByName(name)
	if err != nil {
		return
	}

	uuid = response.AgentUUID
	return
}
