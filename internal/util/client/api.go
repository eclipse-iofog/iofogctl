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
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

// InvalidateCache will clear the cache
func InvalidateCache() {
	pkg.clientReqChan <- ""
	pkg.agentReqChan <- ""
}

// NewControllerClient will return cached client or create new client and cache it
func NewControllerClient(namespace string) (*client.Client, error) {
	pkg.clientReqChan <- namespace
	result := <-pkg.clientChan
	return result.get()
}

// SyncAgentInfo will synchronize local Agent info with backend Agent info
func SyncAgentInfo(namespace string) error {
	var done chan error
	pkg.once.Do(func() {
		done := make(chan error)
		done <- syncAgentInfo(namespace)
	})
	if done != nil {
		return <-done
	}
	return nil
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
	pkg.agentReqChan <- namespace
	result := <-pkg.agentChan
	return result.get()
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
