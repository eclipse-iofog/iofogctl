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

package client

import (
	"fmt"
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

// InvalidateCache will clear the cache
func InvalidateCache() {
	pkg.clientCacheRequestChan <- newClientCacheRequest("")
	pkg.agentCacheRequestChan <- newAgentCacheRequest("")
}

// NewControllerClient will return cached client or create new client and cache it
func NewControllerClient(namespace string) (*client.Client, error) {
	request := newClientCacheRequest(namespace)
	pkg.clientCacheRequestChan <- request
	result := <-request.resultChan
	return result.get()
}

// GetBackendAgents will return cached list of agents or create new list and cache it
func GetBackendAgents(namespace string) ([]client.AgentInfo, error) {
	request := newAgentCacheRequest(namespace)
	pkg.agentCacheRequestChan <- request
	result := <-request.resultChan
	return result.get()
}

// SyncAgentInfo will synchronize local Agent info with backend Agent info
func SyncAgentInfo(namespace string) error {
	request := newAgentSyncRequest(namespace)
	pkg.agentSyncRequestChan <- request
	return <-request.resultChan
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

func GetMicroserviceName(namespace, appName, msvcName string) (name string, err error) {
	clt, err := NewControllerClient(namespace)
	if err != nil {
		return
	}

	response, err := clt.GetMicroserviceByName(appName, msvcName)
	if err != nil {
		return
	}

	name = response.Name
	return
}

func ParseFQName(fqName, resourceType string) (appName, name string, err error) {
	if fqName == "" {
		return "", "", util.NewInputError(fmt.Sprintf("Invalid %s name %s", resourceType, fqName))
	}
	splittedName := strings.Split(fqName, "/")
	switch len(splittedName) {
	case 1:
		if err := util.IsLowerAlphanumeric(resourceType, splittedName[0]); err != nil {
			return "", "", err
		}
		return "", splittedName[0], nil
	case 2:
		if err := util.IsLowerAlphanumeric("application", splittedName[0]); err != nil {
			return "", "", err
		}
		if err := util.IsLowerAlphanumeric(resourceType, splittedName[1]); err != nil {
			return "", "", err
		}
		return splittedName[0], splittedName[1], nil
	default:
		return "", "", util.NewInputError(fmt.Sprintf("Invalid %s name %s", resourceType, fqName))
	}
}

func GetMicroserviceUUID(namespace, appName, name string) (uuid string, err error) {
	clt, err := NewControllerClient(namespace)
	if err != nil {
		return
	}

	response, err := clt.GetMicroserviceByName(appName, name)
	if err != nil {
		return
	}

	uuid = response.UUID
	return
}

func GetAgentConfig(agentName, namespace string) (agentConfig rsc.AgentConfiguration, tags *[]string, agentStatus rsc.AgentStatus, err error) {
	// Try to get agent info directly from controller by name with auth retry
	var agentInfo *client.AgentInfo
	err = ExecuteWithAuthRetry(namespace, func(ctrlClient *client.Client) error {
		var err error
		agentInfo, err = ctrlClient.GetAgentByName(agentName)
		return err
	})
	if err != nil {
		// If agent not found in controller, try local cache approach
		ns, err2 := config.GetNamespace(namespace)
		if err2 != nil {
			err = err2
			return
		}
		// Get config from local cache
		agent, err2 := ns.GetAgent(agentName)
		if err2 != nil {
			err = err2
			return
		}

		// Try to get agent info by UUID with auth retry
		err = ExecuteWithAuthRetry(namespace, func(ctrlClient *client.Client) error {
			var err error
			agentInfo, err = ctrlClient.GetAgentByID(agent.GetUUID())
			if err != nil {
				// The agents might not be provisioned with Controller
				// TODO: Standardize error check and error message here
				if strings.Contains(err.Error(), "NotFoundError") {
					return util.NewInputError("Cannot describe an Agent that is not provisioned with the Controller in Namespace " + namespace)
				}
				return err
			}
			return nil
		})
		if err != nil {
			return
		}
	}

	tags = agentInfo.Tags

	// Get all agents for mapping uuid to name if required with auth retry
	var getAgentList client.ListAgentsResponse
	err = ExecuteWithAuthRetry(namespace, func(ctrlClient *client.Client) error {
		var err error
		getAgentList, err = ctrlClient.ListAgents(client.ListAgentsRequest{})
		return err
	})
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
			upstreamRouters = append(upstreamRouters, getRouterAgentNameFromUUID(agentMapByUUID, upstreamRouterAgentUUID))
		}
		upstreamRoutersPtr = &upstreamRouters
	}

	natsConfig := client.NatsConfig{
		NatsMode:          agentInfo.NatsMode,
		NatsServerPort:    agentInfo.NatsServerPort,
		NatsLeafPort:      agentInfo.NatsLeafPort,
		NatsClusterPort:   agentInfo.NatsClusterPort,
		NatsMqttPort:      agentInfo.NatsMqttPort,
		NatsHttpPort:      agentInfo.NatsHttpPort,
		JsStorageSize:     agentInfo.JsStorageSize,
		JsMemoryStoreSize: agentInfo.JsMemoryStoreSize,
	}

	var upstreamNatsServersPtr *[]string
	if agentInfo.UpstreamNatsServers != nil {
		upstreamNatsServers := []string{}
		for _, upstreamNatsServerUUID := range *agentInfo.UpstreamNatsServers {
			upstreamNatsServers = append(upstreamNatsServers, getNatsAgentNameFromUUID(agentMapByUUID, upstreamNatsServerUUID))
		}
		upstreamNatsServersPtr = &upstreamNatsServers
	}

	var networkRouterPtr *string
	if agentInfo.NetworkRouter != nil {
		networkRouter := getRouterAgentNameFromUUID(agentMapByUUID, *agentInfo.NetworkRouter)
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
			NetworkInterface:          &agentInfo.NetworkInterface,
			DockerURL:                 &agentInfo.DockerURL,
			ContainerEngine:           &agentInfo.ContainerEngine,
			DeploymentType:            &agentInfo.DeploymentType,
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
			GpsMode:                   &agentInfo.GpsMode,
			GpsScanFrequency:          &agentInfo.GpsScanFrequency,
			GpsDevice:                 &agentInfo.GpsDevice,
			EdgeGuardFrequency:        &agentInfo.EdgeGuardFrequency,
			AbstractedHardwareEnabled: &agentInfo.AbstractedHardwareEnabled,
			LogLevel:                  agentInfo.LogLevel,
			DockerPruningFrequency:    agentInfo.DockerPruningFrequency,
			AvailableDiskThreshold:    agentInfo.AvailableDiskThreshold,
			UpstreamRouters:           upstreamRoutersPtr,
			NetworkRouter:             networkRouterPtr,
			RouterConfig:              routerConfig,
			UpstreamNatsServers:       upstreamNatsServersPtr,
			NatsConfig:                natsConfig,
			TimeZone:                  agentInfo.TimeZone,
		},
	}

	agentStatus = rsc.AgentStatus{
		LastActive:            agentInfo.LastActive,
		DaemonStatus:          agentInfo.DaemonStatus,
		SecurityStatus:        agentInfo.SecurityStatus,
		SecurityViolationInfo: agentInfo.SecurityViolationInfo,
		WarningMessage:        agentInfo.WarningMessage,
		UptimeMs:              agentInfo.UptimeMs,
		MemoryUsage:           agentInfo.MemoryUsage,
		DiskUsage:             agentInfo.DiskUsage,
		CPUUsage:              agentInfo.CPUUsage,
		SystemAvailableMemory: agentInfo.SystemAvailableMemory,
		SystemAvailableDisk:   agentInfo.SystemAvailableDisk,
		SystemTotalCPU:        agentInfo.SystemTotalCPU,
		MemoryViolation:       agentInfo.MemoryViolation,
		DiskViolation:         agentInfo.DiskViolation,
		CPUViolation:          agentInfo.CPUViolation,
		RepositoryStatus:      agentInfo.RepositoryStatus,
		LastStatusTimeMsUTC:   agentInfo.LastStatusTimeMsUTC,
		IPAddress:             agentInfo.IPAddress,
		IPAddressExternal:     agentInfo.IPAddressExternal,
		ProcessedMessaged:     agentInfo.ProcessedMessaged,
		MessageSpeed:          agentInfo.MessageSpeed,
		LastCommandTimeMsUTC:  agentInfo.LastCommandTimeMsUTC,
		Version:               agentInfo.Version,
		IsReadyToUpgrade:      agentInfo.IsReadyToUpgrade,
		IsReadyToRollback:     agentInfo.IsReadyToRollback,
		Tunnel:                agentInfo.Tunnel,
		VolumeMounts:          convertVolumeMounts(agentInfo.VolumeMounts),
		GpsStatus:             agentInfo.GpsStatus,
	}

	return agentConfig, tags, agentStatus, err
}

func convertVolumeMounts(volumeMountInfos []client.VolumeMountInfo) []rsc.VolumeMount {
	var volumeMounts []rsc.VolumeMount
	for _, vm := range volumeMountInfos {
		volumeMounts = append(volumeMounts, rsc.VolumeMount{
			Name:          vm.Name,
			UUID:          vm.UUID,
			ConfigMapName: vm.ConfigMapName,
			SecretName:    vm.SecretName,
			Version:       vm.Version,
		})
	}
	return volumeMounts
}
