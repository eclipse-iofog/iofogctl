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
	"fmt"
	"strings"
	"sync"

	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

var clientCache map[string]*client.Client
var agentCache map[string][]client.AgentInfo
var mux sync.Mutex

func init() {
	InvalidateCache()
}

func InvalidateCache() {
	mux.Lock()
	defer mux.Unlock()

	clientCache = make(map[string]*client.Client)
	agentCache = make(map[string][]client.AgentInfo)
}

func NewControllerClient(namespace string) (*client.Client, error) {
	mux.Lock()
	defer mux.Unlock()

	if cachedClient, exists := clientCache[namespace]; exists {
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
	cachedClient, err := client.NewAndLogin(client.Options{Endpoint: endpoint}, user.Email, user.GetRawPassword())
	if err != nil {
		return nil, err
	}
	clientCache[namespace] = cachedClient

	return cachedClient, nil
}

func IsEdgeResourceCapable(namespace string) error {
	// Check Controller API handles edge resources
	clt, err := NewControllerClient(namespace)
	if err != nil {
		return err
	}
	if err := clt.IsEdgeResourceCapable(); err != nil {
		return err
	}
	return nil
}

func GetBackendAgents(namespace string) ([]client.AgentInfo, error) {
	if cachedAgents, exist := agentCache[namespace]; exist {
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
	agentCache[namespace] = agentList.Agents // TODO: Fix concurrent writes issue here
	return agentList.Agents, nil
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

	uuid = response.UUID
	return
}

func GetAgentConfig(agentName, namespace string) (agentConfig rsc.AgentConfiguration, tags *[]string, err error) {
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return
	}
	// Get config
	agent, err := ns.GetAgent(agentName)
	if err != nil {
		return
	}

	// Connect to controller
	ctrl, err := NewControllerClient(namespace)
	if err != nil {
		return
	}

	agentInfo, err := ctrl.GetAgentByID(agent.GetUUID())
	if err != nil {
		// The agents might not be provisioned with Controller
		// TODO: Standardize error check and error message here
		if strings.Contains(err.Error(), "NotFoundError") {
			err = util.NewInputError("Cannot describe an Agent that is not provisioned with the Controller in Namespace " + namespace)
			return
		}
		return
	}
	tags = agentInfo.Tags

	// Get all agents for mapping uuid to name if required
	getAgentList, err := ctrl.ListAgents(client.ListAgentsRequest{})
	if err != nil {
		return
	}
	// Map by uuid for easier access
	agentMapByUUID := make(map[string]client.AgentInfo)
	for idx := range getAgentList.Agents {
		agent := &getAgentList.Agents[idx]
		agentMapByUUID[agent.UUID] = *agent
	}

	fogType, found := rsc.FogTypeIntMap[agentInfo.FogType]
	if !found {
		fogType = "auto"
	}

	routerConfig := client.RouterConfig{
		RouterMode:      &agentInfo.RouterMode,
		MessagingPort:   agentInfo.MessagingPort,
		EdgeRouterPort:  agentInfo.EdgeRouterPort,
		InterRouterPort: agentInfo.InterRouterPort,
	}

	var upstreamRoutersPtr *[]string

	if agentInfo.UpstreamRouters != nil {
		upstreamRouters := []string{}
		for _, upstreamRouterAgentUUID := range *agentInfo.UpstreamRouters {
			upstreamRouters = append(upstreamRouters, getAgentNameFromUUID(agentMapByUUID, upstreamRouterAgentUUID))
		}
		upstreamRoutersPtr = &upstreamRouters
	}

	var networkRouterPtr *string
	if agentInfo.NetworkRouter != nil {
		networkRouter := getAgentNameFromUUID(agentMapByUUID, *agentInfo.NetworkRouter)
		networkRouterPtr = &networkRouter
	}

	agentConfig = rsc.AgentConfiguration{
		Name:        agentInfo.Name,
		Location:    agentInfo.Location,
		Latitude:    agentInfo.Latitude,
		Longitude:   agentInfo.Longitude,
		Description: agentInfo.Description,
		FogType:     &fogType,
		AgentConfiguration: client.AgentConfiguration{
			DockerURL:                 &agentInfo.DockerURL,
			DiskLimit:                 &agentInfo.DiskLimit,
			DiskDirectory:             &agentInfo.DiskDirectory,
			MemoryLimit:               &agentInfo.MemoryLimit,
			CPULimit:                  &agentInfo.CPULimit,
			LogLimit:                  &agentInfo.LogLimit,
			LogDirectory:              &agentInfo.LogDirectory,
			LogFileCount:              &agentInfo.LogFileCount,
			StatusFrequency:           &agentInfo.StatusFrequency,
			ChangeFrequency:           &agentInfo.ChangeFrequency,
			DeviceScanFrequency:       &agentInfo.DeviceScanFrequency,
			BluetoothEnabled:          &agentInfo.BluetoothEnabled,
			WatchdogEnabled:           &agentInfo.WatchdogEnabled,
			AbstractedHardwareEnabled: &agentInfo.AbstractedHardwareEnabled,
			LogLevel:                  agentInfo.LogLevel,
			DockerPruningFrequency:    agentInfo.DockerPruningFrequency,
			AvailableDiskThreshold:    agentInfo.AvailableDiskThreshold,
			UpstreamRouters:           upstreamRoutersPtr,
			NetworkRouter:             networkRouterPtr,
			RouterConfig:              routerConfig,
		},
	}

	return agentConfig, tags, err
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
