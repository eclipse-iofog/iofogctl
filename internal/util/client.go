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

	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

var clientCache map[string]*client.Client
var agentCache map[string][]client.AgentInfo
var clientReqChan chan string
var clientChan chan clientCacheResult
var agentReqChan chan string
var agentChan chan agentCacheResult
var agentConfigReqChan chan string
var agentConfigChan chan error

func init() {
	clientReqChan = make(chan string, 10)
	clientChan = make(chan clientCacheResult)
	agentReqChan = make(chan string, 10)
	agentChan = make(chan agentCacheResult)
	agentConfigReqChan = make(chan string, 10)
	agentConfigChan = make(chan error)
	go clientRoutine()
	go agentRoutine()
	go agentConfigRoutine()
	InvalidateCache()
}

type clientCacheResult struct {
	err    error
	client *client.Client
}

func (ccr *clientCacheResult) get() (*client.Client, error) {
	return ccr.client, ccr.err
}

type agentCacheResult struct {
	err    error
	agents []client.AgentInfo
}

func (acr *agentCacheResult) get() ([]client.AgentInfo, error) {
	return acr.agents, acr.err
}

func InvalidateCache() {
	clientReqChan <- ""
	agentReqChan <- ""
}

func clientRoutine() {
	for {
		namespace := <-clientReqChan
		// Invalidate cache
		if namespace == "" {
			clientCache = make(map[string]*client.Client)
			continue
		}
		result := clientCacheResult{}
		// From cache
		if cachedClient, exists := clientCache[namespace]; exists {
			result.client = cachedClient
			clientChan <- result
			continue
		}
		// Create new client
		ioClient, err := newControllerClient(namespace)
		// Failure
		if err != nil {
			result.err = err
			clientChan <- result
			continue
		}
		// Save to cache and return new client
		clientCache[namespace] = ioClient
		result.client = ioClient
		clientChan <- result
	}
}

func agentRoutine() {
	for {
		namespace := <-agentReqChan
		if namespace == "" {
			// Invalidate cache
			agentCache = make(map[string][]client.AgentInfo)
			continue
		}
		result := agentCacheResult{}
		// From cache
		if cachedAgents, exist := agentCache[namespace]; exist {
			result.agents = cachedAgents
			agentChan <- result
			continue
		}
		// Client to get agents
		ioClient, err := NewControllerClient(namespace)
		if err != nil {
			result.err = err
			agentChan <- result
			continue
		}
		// Get agents
		agents, err := getBackendAgents(namespace, ioClient)
		if err != nil {
			result.err = err
			agentChan <- result
			continue
		}
		// Save to cache and return new agents
		agentCache[namespace] = agents
		result.agents = agents
		agentChan <- result
	}
}

func NewControllerClient(namespace string) (*client.Client, error) {
	clientReqChan <- namespace
	result := <-clientChan
	return result.get()
}

func UpdateAgentCache(namespace string) error {
	agentConfigReqChan <- namespace
	return <-agentConfigChan
}

func updateAgentCache(namespace string) error {
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
			agentConfigChan <- err
			continue
		}
	}

	if localAgent != nil {
		if err := ns.AddAgent(localAgent); err != nil {
			return err
		}
	}

	return config.Flush()
}

func agentConfigRoutine() {
	complete := false
	for {
		namespace := <-agentConfigReqChan
		if complete {
			agentConfigChan <- nil
			continue
		}
		if err := updateAgentCache(namespace); err != nil {
			agentConfigChan <- err
			continue
		}
		complete = true
		agentConfigChan <- nil
	}
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
	agentReqChan <- namespace
	result := <-agentChan
	return result.get()
}

func getBackendAgents(namespace string, ioClient *client.Client) ([]client.AgentInfo, error) {
	agentList, err := ioClient.ListAgents(client.ListAgentsRequest{})
	if err != nil {
		return nil, err
	}
	agentCache[namespace] = agentList.Agents
	return agentList.Agents, nil
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
