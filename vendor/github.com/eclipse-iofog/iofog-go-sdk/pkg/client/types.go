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

// Registries

type RegistryInfo struct {
	ID           int    `json:"id"`
	URL          string `json:"url"`
	IsPublic     bool   `json:"isPublic"`
	IsSecure     bool   `json:"isSecure"`
	Certificate  string `json:"certificate"`
	RequiresCert bool   `json:"requiresCert"`
	Username     string `json:"username"`
	Email        string `json:"userEmail"`
}

type RegistryCreateRequest struct {
	URL          string `json:"url"`
	IsPublic     bool   `json:"isPublic"`
	Certificate  string `json:"certificate"`
	RequiresCert bool   `json:"requiresCert"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	Password     string `json:"password"`
}

type RegistryCreateResponse struct {
	ID int `json:"id"`
}

type RegistryUpdateRequest struct {
	URL          *string `json:"url,omitempty"`
	IsPublic     *bool   `json:"isPublic,omitempty"`
	Certificate  *string `json:"certificate,omitempty"`
	RequiresCert *bool   `json:"requiresCert,omitempty"`
	Username     *string `json:"username,omitempty"`
	Email        *string `json:"email,omitempty"`
	Password     *string `json:"password,omitempty"`
	ID           int     `json:"-"`
}

type RegistryListResponse struct {
	Registries []RegistryInfo `json:"registries"`
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

type MicroserviceStatus struct {
	Status            string  `json:"status"`
	StartTimne        int64   `json:"startTime"`
	OperatingDuration int64   `json:"operatingDuration"`
	MemoryUsage       float64 `json:"memoryUsage"`
	CpuUsage          float64 `json:"cpuUsage"`
	ContainerId       string  `json:"containerId"`
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
	Commands          []string                    `json:"cmd"`
	Env               []MicroserviceEnvironment   `json:"env"`
	Status            MicroserviceStatus          `json:"status"`
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
	RegistryID     int                         `json:"registryId"`
	Ports          []MicroservicePortMapping   `json:"ports"`
	Volumes        []MicroserviceVolumeMapping `json:"volumeMappings"`
	Routes         []string                    `json:"routes"`
	Commands       []string                    `json:"cmd,omitempty"`
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
	RegistryID        *int                         `json:"registryId"`
	Ports             []MicroservicePortMapping    `json:"-"` // Ports are not valid in Controller PATCH call, need to use separate API calls
	Volumes           *[]MicroserviceVolumeMapping `json:"volumeMappings,omitempty"`
	Commands          *[]string                    `json:"cmd,omitempty"`
	Routes            []string                     `json:"-"` // Routes are not valid in Controller PATCH call, need to use separate API calls
	Env               *[]MicroserviceEnvironment   `json:"env,omitempty"`
	Images            []CatalogImage               `json:"images"`
	Rebuild           bool                         `json:"rebuild"`
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
	UUID                      string  `json:"uuid" yaml:"uuid"`
	Name                      string  `json:"name" yaml:"name"`
	Location                  string  `json:"location" yaml:"location"`
	Latitude                  float64 `json:"latitude" yaml:"latitude"`
	Longitude                 float64 `json:"longitude" yaml:"longitude"`
	Description               string  `json:"description" yaml:"description"`
	DockerURL                 string  `json:"dockerUrl" yaml:"dockerUrl"`
	DiskLimit                 int64   `json:"diskLimit" yaml:"diskLimit"`
	DiskDirectory             string  `json:"diskDirectory" yaml:"diskDirectory"`
	MemoryLimit               int64   `json:"memoryLimit" yaml:"memoryLimit"`
	CPULimit                  int64   `json:"cpuLimit" yaml:"cpuLimit"`
	LogLimit                  int64   `json:"logLimit" yaml:"logLimit"`
	LogDirectory              string  `json:"logDirectory" yaml:"logDirectory"`
	LogFileCount              int64   `json:"logFileCount" yaml:"logFileCount"`
	StatusFrequency           float64 `json:"statusFrequency" yaml:"statusFrequency"`
	ChangeFrequency           float64 `json:"changeFrequency" yaml:"changeFrequency"`
	DeviceScanFrequency       float64 `json:"deviceScanFrequency" yaml:"deviceScanFrequency"`
	BluetoothEnabled          bool    `json:"bluetoothEnabled" yaml:"bluetoothEnabled"`
	WatchdogEnabled           bool    `json:"watchdogEnabled" yaml:"watchdogEnabled"`
	AbstractedHardwareEnabled bool    `json:"abstractedHardwareEnabled" yaml:"abstractedHardwareEnabled"`
	CreatedTimeRFC3339        string  `json:"created_at" yaml:"created"`
	UpdatedTimeRFC3339        string  `json:"updated_at" yaml:"updated"`
	LastActive                int64   `json:"lastActive" yaml:"lastActive"`
	DaemonStatus              string  `json:"daemonStatus" yaml:"daemonStatus"`
	UptimeMs                  int64   `json:"daemonOperatingDuration" yaml:"uptime"`
	MemoryUsage               float64 `json:"memoryUsage" yaml:"memoryUsage"`
	DiskUsage                 float64 `json:"diskUsage" yaml:"diskUsage"`
	CPUUsage                  float64 `json:"cpuUsage" yaml:"cpuUsage"`
	MemoryViolation           string  `json:"memoryViolation" yaml:"memoryViolation"`
	DiskViolation             string  `json:"diskViolation" yaml:"diskViolation"`
	CPUViolation              string  `json:"cpuViolation" yaml:"cpuViolation"`
	MicroserviceStatus        string  `json:"microserviceStatus" yaml:"microserviceStatus"`
	RepositoryCount           int64   `json:"repositoryCount" yaml:"repositoryCount"`
	RepositoryStatus          string  `json:"repositoryStatus" yaml:"repositoryStatus"`
	LastStatusTimeMsUTC       int64   `json:"lastStatusTime" yaml:"lastStatusTime"`
	IPAddress                 string  `json:"ipAddress" yaml:"ipAddress"`
	IPAddressExternal         string  `json:"ipAddressExternal" yaml:"ipAddressExternal"`
	ProcessedMessaged         int64   `json:"processedMessages" yaml:"ProcessedMessages"`
	MicroserviceMessageCount  int64   `json:"microserviceMessageCounts" yaml:"microserviceMessageCount"`
	MessageSpeed              float64 `json:"messageSpeed" yaml:"messageSpeed"`
	LastCommandTimeMsUTC      int64   `json:"lastCommandTime" yaml:"lastCommandTime"`
	NetworkInterface          string  `json:"networkInterface" yaml:"networkInterface"`
	Version                   string  `json:"version" yaml:"version"`
	IsReadyToUpgrade          bool    `json:"isReadyToUpgrade" yaml:"isReadyToUpgrade"`
	IsReadyToRollback         bool    `json:"isReadyToRollback" yaml:"isReadyToRollback"`
	Tunnel                    string  `json:"tunnel" yaml:"tunnel"`
	FogType                   int     `json:"fogTypeId" yaml:"fogTypeId"`
}

type AgentConfiguration struct {
	DockerURL                 *string  `json:"dockerUrl,omitempty" yaml:"dockerUrl"`
	DiskLimit                 *int64   `json:"diskLimit,omitempty" yaml:"diskLimit"`
	DiskDirectory             *string  `json:"diskDirectory,omitempty" yaml:"diskDirectory"`
	MemoryLimit               *int64   `json:"memoryLimit,omitempty" yaml:"memoryLimit"`
	CPULimit                  *int64   `json:"cpuLimit,omitempty" yaml:"cpuLimit"`
	LogLimit                  *int64   `json:"logLimit,omitempty" yaml:"logLimit"`
	LogDirectory              *string  `json:"logDirectory,omitempty" yaml:"logDirectory"`
	LogFileCount              *int64   `json:"logFileCount,omitempty" yaml:"logFileCount"`
	StatusFrequency           *float64 `json:"statusFrequency,omitempty" yaml:"statusFrequency"`
	ChangeFrequency           *float64 `json:"changeFrequency,omitempty" yaml:"changeFrequency"`
	DeviceScanFrequency       *float64 `json:"deviceScanFrequency,omitempty" yaml:"deviceScanFrequency"`
	BluetoothEnabled          *bool    `json:"bluetoothEnabled,omitempty" yaml:"bluetoothEnabled"`
	WatchdogEnabled           *bool    `json:"watchdogEnabled,omitempty" yaml:"watchdogEnabled"`
	AbstractedHardwareEnabled *bool    `json:"abstractedHardwareEnabled,omitempty" yaml:"abstractedHardwareEnabled"`
}

type AgentUpdateRequest struct {
	UUID        string  `json:"-"`
	Name        string  `json:"name,omitempty" yaml:"name"`
	Location    string  `json:"location,omitempty" yaml:"location"`
	Latitude    float64 `json:"latitude,omitempty" yaml:"latitude"`
	Longitude   float64 `json:"longitude,omitempty" yaml:"longitude"`
	Description string  `json:"description,omitempty" yaml:"description"`
	FogType     int64   `json:"fogType,omitempty" yaml:"agentType"`
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
