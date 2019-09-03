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

package client

// Flows

type FlowInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActivated bool   `json:"isActivated"`
	UserID      int    `json:"userId"`
	ID          int    `json:"id"`
}

type FlowCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type FlowCreateResponse struct {
	ID int `json:"id"`
}

type FlowUpdateRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	IsActivated *bool   `json:"isActivated,omitempty"`
	ID          int     `json:"-"`
}

type FlowListResponse struct {
	Flows []FlowInfo `json:"flows"`
}

// Catalog (Keeping it basic, because it will be reworked soon)

type CatalogImage struct {
	ContainerImage string `json:"containerImage"`
	AgentTypeID    int    `json:"fogTypeId"`
}

type CatalogItemInfo struct {
	ID          int            `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Images      []CatalogImage `json:"images"`
	RegistryID  int            `json:"registryId"`
	Category    string         `json:"category"`
}

type CatalogItemCreateRequest struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Images      []CatalogImage `json:"images"`
	RegistryID  int            `json:"registryId"`
}

type CatalogItemCreateResponse struct {
	ID int `json:"id"`
}

type CatalogItemUpdateRequest struct {
	ID          int
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	Images      []CatalogImage `json:"images,omitempty"`
	RegistryID  int            `json:"registryId,omitempty"`
}

type CatalogListResponse struct {
	CatalogItems []CatalogItemInfo `json:"catalogItems"`
}

// Microservices

type MicroservicePortMapping struct {
	Internal   int  `json:"internal"`
	External   int  `json:"external"`
	PublicMode bool `json:"publicMode"`
}

type MicroserviceVolumeMapping struct {
	HostDestination      string `json:"hostDestination"`
	ContainerDestination string `json:"containerDestination"`
	AccessMode           string `json:"accessMode"`
}

type MicroserviceEnvironment struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type MicroserviceInfo struct {
	UUID              string                      `json:"uuid"`
	Config            string                      `json:"config"`
	Name              string                      `json:"name"`
	RootHostAccess    bool                        `json:"rootHostAccess"`
	LogSize           int                         `json:"logSize"`
	Delete            bool                        `json:"delete"`
	DeleteWithCleanup bool                        `json:"deleteWithCleanup"`
	FlowID            int                         `json:"flowId"`
	CatalogItemID     int                         `json:"catalogItemId"`
	AgentUUID         string                      `json:"iofogUuid"`
	UserID            int                         `json:"userId"`
	RegistryID        int                         `json:"registryId"`
	Ports             []MicroservicePortMapping   `json:"ports"`
	Volumes           []MicroserviceVolumeMapping `json:"volumeMappings"`
	Routes            []string                    `json:"routes"`
	Env               []MicroserviceEnvironment   `json:"env"`
	Images            []CatalogImage              `json:"images"`
}

type MicroserviceCreateRequest struct {
	Config         string                      `json:"config"`
	Name           string                      `json:"name"`
	RootHostAccess bool                        `json:"rootHostAccess"`
	LogSize        int                         `json:"logSize"`
	FlowID         int                         `json:"flowId"`
	CatalogItemID  int                         `json:"catalogItemId,omitempty"`
	AgentUUID      string                      `json:"iofogUuid"`
	Ports          []MicroservicePortMapping   `json:"ports"`
	Volumes        []MicroserviceVolumeMapping `json:"volumeMappings"`
	Routes         []string                    `json:"routes"`
	Env            []MicroserviceEnvironment   `json:"env"`
	Images         []CatalogImage              `json:"images,omitempty"`
}

type MicroserviceUpdateRequest struct {
	UUID              string                       `json:"-"`
	Config            *string                      `json:"config,omitempty"`
	Name              *string                      `json:"name,omitempty"`
	RootHostAccess    *bool                        `json:"rootHostAccess,omitempty"`
	LogSize           *int                         `json:"logSize,omitempty"`
	Delete            *bool                        `json:"delete,omitempty"`
	DeleteWithCleanup *bool                        `json:"deleteWithCleanup,omitempty"`
	FlowID            *int                         `json:"flowId,omitempty"`
	AgentUUID         *string                      `json:"iofogUuid,omitempty"`
	UserID            *int                         `json:"userId,omitempty"`
	Ports             []MicroservicePortMapping    `json:"-"` // Ports are not valid in Controller PATCH call, need to use separate API calls
	Volumes           *[]MicroserviceVolumeMapping `json:"volumeMappings,omitempty"`
	Routes            []string                     `json:"-"` // Routes are not valid in Controller PATCH call, need to use separate API calls
	Env               []MicroserviceEnvironment    `json:"env,omitempty"`
	Images            []CatalogImage               `json:"images,omitempty"`
}

type MicroserviceCreateResponse struct {
	UUID string `json:"uuid"`
}

type MicroserviceListResponse struct {
	Microservices []MicroserviceInfo
}

type MicroservicePortMappingListResponse struct {
	PortMappings []MicroservicePortMapping `json:"ports"`
}

// Users

type User struct {
	Name     string `json:"firstName"`
	Surname  string `json:"lastName"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ControllerStatus struct {
	Status        string  `json:"status"`
	UptimeSeconds float64 `json:"uptimeSec"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string `json:"accessToken"`
}

type CreateAgentRequest struct {
	Name    string `json:"name"`
	FogType int32  `json:"fogType"`
}

type CreateAgentResponse struct {
	UUID string
}

type GetAgentProvisionKeyResponse struct {
	Key             string `json:"key"`
	ExpireTimeMsUTC int64  `json:"expirationTime"`
}

type AgentInfo struct {
	UUID                      string  `json:"uuid" yml:"uuid"`
	Name                      string  `json:"name" yml:"name"`
	Location                  string  `json:"location" yml:"location"`
	Latitude                  float64 `json:"latitude" yml:"latitude"`
	Longitude                 float64 `json:"longitude" yml:"longitude"`
	Description               string  `json:"description" yml:"description"`
	DockerURL                 string  `json:"dockerUrl" yml:"dockerUrl"`
	DiskLimit                 int64   `json:"diskLimit" yml:"diskLimit"`
	DiskDirectory             string  `json:"diskDirectory" yml:"diskDirectory"`
	MemoryLimit               int64   `json:"memoryLimit" yml:"memoryLimit"`
	CPULimit                  int64   `json:"cpuLimit" yml:"cpuLimit"`
	LogLimit                  int64   `json:"logLimit" yml:"logLimit"`
	LogDirectory              string  `json:"logDirectory" yml:"logDirectory"`
	LogFileCount              int64   `json:"logFileCount" yml:"logFileCount"`
	StatusFrequency           float64 `json:"statusFrequency" yml:"statusFrequency"`
	ChangeFrequency           float64 `json:"changeFrequency" yml:"changeFrequency"`
	DeviceScanFrequency       float64 `json:"deviceScanFrequency" yml:"deviceScanFrequency"`
	BluetoothEnabled          bool    `json:"bluetoothEnabled" yml:"bluetoothEnabled"`
	WatchdogEnabled           bool    `json:"watchdogEnabled" yml:"watchdogEnabled"`
	AbstractedHardwareEnabled bool    `json:"abstractedHardwareEnabled" yml:"abstractedHardwareEnabled"`
	CreatedTimeRFC3339        string  `json:"created_at" yml:"created"`
	UpdatedTimeRFC3339        string  `json:"updated_at" yml:"updated"`
	LastActive                int64   `json:"lastActive" yml:"lastActive"`
	DaemonStatus              string  `json:"daemonStatus" yml:"daemonStatus"`
	UptimeMs                  int64   `json:"daemonOperatingDuration" yml:"uptime"`
	MemoryUsage               float64 `json:"memoryUsage" yml:"memoryUsage"`
	DiskUsage                 float64 `json:"diskUsage" yml:"diskUsage"`
	CPUUsage                  float64 `json:"cpuUsage" yml:"cpuUsage"`
	MemoryViolation           string  `json:"memoryViolation" yml:"memoryViolation"`
	DiskViolation             string  `json:"diskViolation" yml:"diskViolation"`
	CPUViolation              string  `json:"cpuViolation" yml:"cpuViolation"`
	MicroserviceStatus        string  `json:"microserviceStatus" yml:"microserviceStatus"`
	RepositoryCount           int64   `json:"repositoryCount" yml:"repositoryCount"`
	RepositoryStatus          string  `json:"repositoryStatus" yml:"repositoryStatus"`
	LastStatusTimeMsUTC       int64   `json:"lastStatusTime" yml:"LastStatusTime"`
	IPAddress                 string  `json:"ipAddress" yml:"ipAddress"`
	IPAddressExternal         string  `json:"ipAddressExternal" yml:"ipAddressExternal"`
	ProcessedMessaged         int64   `json:"processedMessages" yml:"ProcessedMessages"`
	MicroserviceMessageCount  int64   `json:"microserviceMessageCounts" yml:"microserviceMessageCount"`
	MessageSpeed              float64 `json:"messageSpeed" yml:"messageSpeed"`
	LastCommandTimeMsUTC      int64   `json:"lastCommandTime" yml:"lastCommandTime"`
	NetworkInterface          string  `json:"networkInterface" yml:"networkInterface"`
	Version                   string  `json:"version" yml:"version"`
	IsReadyToUpgrade          bool    `json:"isReadyToUpgrade" yml:"isReadyToUpgrade"`
	IsReadyToRollback         bool    `json:"isReadyToRollback" yml:"isReadyToRollback"`
	Tunnel                    string  `json:"tunnel" yml:"tunnel"`
}

type AgentConfiguration struct {
	DockerURL                 *string  `json:"dockerUrl,omitempty"`
	DiskLimit                 *int64   `json:"diskLimit,omitempty"`
	DiskDirectory             *string  `json:"diskDirectory,omitempty"`
	MemoryLimit               *int64   `json:"memoryLimit,omitempty"`
	CPULimit                  *int64   `json:"cpuLimit,omitempty"`
	LogLimit                  *int64   `json:"logLimit,omitempty"`
	LogDirectory              *string  `json:"logDirectory,omitempty"`
	LogFileCount              *int64   `json:"logFileCount,omitempty"`
	StatusFrequency           *float64 `json:"statusFrequency,omitempty"`
	ChangeFrequency           *float64 `json:"changeFrequency,omitempty"`
	DeviceScanFrequency       *float64 `json:"deviceScanFrequency,omitempty"`
	BluetoothEnabled          *bool    `json:"bluetoothEnabled,omitempty"`
	WatchdogEnabled           *bool    `json:"watchdogEnabled,omitempty"`
	AbstractedHardwareEnabled *bool    `json:"abstractedHardwareEnabled,omitempty"`
}

type AgentUpdateRequest struct {
	UUID        string  `json:"-"`
	Name        string  `json:"name,omitempty"`
	Location    string  `json:"location,omitempty"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
	Description string  `json:"description,omitempty"`
	FogType     int64   `json:"fogType,omitempty"`
	AgentConfiguration
}

type ListAgentsResponse struct {
	Agents []AgentInfo `json:"fogs"`
}

type AgentListFilter struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	Condition string `json:"condition"`
}

type ConnectorInfo struct {
	IP      string `json:"publicIp"`
	Name    string `json:"name"`
	Domain  string `json:"domain"`
	DevMode bool   `json:"devMode"`
}

type ConnectorInfoList struct {
	Connectors []ConnectorInfo `json:"connectors"`
}
