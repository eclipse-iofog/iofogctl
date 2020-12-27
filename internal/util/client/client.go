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

package client

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

// clientCacheRoutine handles concurrent requests for a cached Controller client
func clientCacheRoutine() {
	for {
		namespace := <-pkg.clientReqChan
		// Invalidate cache
		if namespace == "" {
			pkg.clientCache = make(map[string]*client.Client)
			continue
		}
		result := clientCacheResult{}
		// From cache
		if cachedClient, exists := pkg.clientCache[namespace]; exists {
			result.client = cachedClient
			pkg.clientChan <- result
			continue
		}
		// Create new client
		ioClient, err := newControllerClient(namespace)
		// Failure
		if err != nil {
			result.err = err
			pkg.clientChan <- result
			continue
		}
		// Save to cache and return new client
		pkg.clientCache[namespace] = ioClient
		result.client = ioClient
		pkg.clientChan <- result
	}
}

// agentCacheRoutine handles concurrent requests for a cached list of Agents
func agentCacheRoutine() {
	for {
		namespace := <-pkg.agentReqChan
		if namespace == "" {
			// Invalidate cache
			pkg.agentCache = make(map[string][]client.AgentInfo)
			continue
		}
		result := agentCacheResult{}
		// From cache
		if cachedAgents, exist := pkg.agentCache[namespace]; exist {
			result.agents = cachedAgents
			pkg.agentChan <- result
			continue
		}
		// Client to get agents
		ioClient, err := NewControllerClient(namespace)
		if err != nil {
			result.err = err
			pkg.agentChan <- result
			continue
		}
		// Get agents
		agents, err := getBackendAgents(namespace, ioClient)
		if err != nil {
			result.err = err
			pkg.agentChan <- result
			continue
		}
		// Save to cache and return new agents
		pkg.agentCache[namespace] = agents
		result.agents = agents
		pkg.agentChan <- result
	}
}

func syncAgentInfo(namespace string) error {
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
	if _, ok := controlPlane.(*rsc.LocalControlPlane); ok {
		// Do not update local Agents
		return nil
	}
	// Generate map of config Agents
	agentsMap := make(map[string]*rsc.RemoteAgent)
	var localAgent *rsc.LocalAgent
	for _, baseAgent := range ns.GetAgents() {
		if v, ok := baseAgent.(*rsc.LocalAgent); ok {
			localAgent = v
		} else {
			agentsMap[baseAgent.GetName()] = baseAgent.(*rsc.RemoteAgent)
		}
	}

	// Get backend Agents
	backendAgents, err := GetBackendAgents(namespace)
	if err != nil {
		return err
	}

	// Generate cache types
	agents := make([]rsc.RemoteAgent, len(backendAgents))
	for idx := range backendAgents {
		backendAgent := &backendAgents[idx]
		if localAgent != nil && backendAgent.Name == localAgent.Name {
			localAgent.UUID = backendAgent.UUID
			continue
		}

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

		agents[idx] = agent
	}

	// Overwrite the Agents
	ns.DeleteAgents()
	for idx := range agents {
		if err := ns.AddAgent(&agents[idx]); err != nil {
			return err
		}
	}

	if localAgent != nil {
		if err := ns.AddAgent(localAgent); err != nil {
			return err
		}
	}

	return config.Flush()
}

func newControllerClient(namespace string) (*client.Client, error) {
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
	cachedClient, err := client.NewAndLogin(client.Options{Endpoint: endpoint}, user.Email, user.GetRawPassword())
	if err != nil {
		return nil, err
	}
	pkg.clientCache[namespace] = cachedClient

	return cachedClient, nil
}

func getBackendAgents(namespace string, ioClient *client.Client) ([]client.AgentInfo, error) {
	agentList, err := ioClient.ListAgents(client.ListAgentsRequest{})
	if err != nil {
		return nil, err
	}
	pkg.agentCache[namespace] = agentList.Agents
	return agentList.Agents, nil
}

func getAgentNameFromUUID(agentMapByUUID map[string]client.AgentInfo, uuid string) (name string) {
	if uuid == iofog.VanillaRouterAgentName {
		return uuid
	}
	agent, found := agentMapByUUID[uuid]
	if !found {
		util.PrintNotify(fmt.Sprintf("Could not find Router: %s\n", uuid))
		name = "UNKNOWN ROUTER: " + uuid
	} else {
		name = agent.Name
	}
	return
}
