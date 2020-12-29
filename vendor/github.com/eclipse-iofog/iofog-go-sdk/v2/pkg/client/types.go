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

// Flows - Keep for legacy
type FlowInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActivated bool   `json:"isActivated"`
	IsSystem    bool   `json:"isSystem"`
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
	IsSystem    *bool   `json:"isSystem,omitempty"`
	ID          int     `json:"-"`
}

type FlowListResponse struct {
	Flows []FlowInfo `json:"flows"`
}

// Applications
type ApplicationInfo struct {
	Name          string             `json:"name"`
	Description   string             `json:"description"`
	IsActivated   bool               `json:"isActivated"`
	IsSystem      bool               `json:"isSystem"`
	UserID        int                `json:"userId"`
	ID            int                `json:"id"`
	Microservices []MicroserviceInfo `json:"microservices"`
	Routes        []Route            `json:"routes"`
}

type ApplicationCreateRequest struct {
	Name          string                           `json:"name"`
	Description   string                           `json:"description,omitempty"`
	Microservices []MicroserviceCreateRequest      `json:"microservices"`
	Routes        *[]ApplicationRouteCreateRequest `json:"routes"`
	Template      *ApplicationTemplate             `json:"template,omitempty" yaml:"template,omitempty" `
}

type ApplicationCreateResponse struct {
	ID int `json:"id"`
}

type ApplicationUpdateRequest struct {
	Name          *string                          `json:"name,omitempty"`
	Description   *string                          `json:"description,omitempty"`
	IsActivated   *bool                            `json:"isActivated,omitempty"`
	IsSystem      *bool                            `json:"isSystem,omitempty"`
	Microservices *[]MicroserviceCreateRequest     `json:"microservices,omitempty"`
	Routes        *[]ApplicationRouteCreateRequest `json:"routes,omitempty"`
	Template      *ApplicationTemplate             `json:"template,omitempty"`
}

type ApplicationPatchRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	IsActivated *bool   `json:"isActivated,omitempty"`
	IsSystem    *bool   `json:"isSystem,omitempty"`
}

type ApplicationListResponse struct {
	Applications []ApplicationInfo `json:"applications"`
}

// Application Templates
type ApplicationTemplate struct {
	Name        string                   `json:"name,omitempty"`
	Description string                   `json:"description,omitempty"`
	Variables   []TemplateVariable       `json:"variables,omitempty"`
	Application *ApplicationTemplateInfo `json:"application,omitempty"`
}

type ApplicationTemplateCreateRequest = ApplicationTemplate

type TemplateVariable struct {
	Key          string      `json:"key" yaml:"key,omitempty"`
	Description  string      `json:"description" yaml:"description,omitempty"`
	DefaultValue interface{} `json:"defaultValue,omitempty" yaml:"defaultValue,omitempty"`
	Value        interface{} `json:"value,omitempty" yaml:"value,omitempty"`
}

type ApplicationTemplateInfo struct {
	Microservices []MicroserviceCreateRequest     `json:"microservices"`
	Routes        []ApplicationRouteCreateRequest `json:"routes"`
}

type ApplicationTemplateCreateResponse struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

type ApplicationTemplateUpdateRequest = ApplicationTemplate

type ApplicationTemplateUpdateResponse = ApplicationTemplateCreateResponse

type ApplicationTemplateMetadataUpdateRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type ApplicationTemplateListResponse struct {
	ApplicationTemplates []ApplicationTemplate
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
	Internal   interface{} `json:"internal"`
	External   interface{} `json:"external"`
	Public     interface{} `json:"publicPort,omitempty"`
	Host       string      `json:"host,omitempty"`
	Protocol   string      `json:"protocol,omitempty"`
	PublicLink string      `json:"publicLink,omitempty"`
}

type MicroserviceVolumeMapping struct {
	HostDestination      string `json:"hostDestination"`
	ContainerDestination string `json:"containerDestination"`
	AccessMode           string `json:"accessMode"`
	Type                 string `json:"type,omitempty"`
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
	CPUUsage          float64 `json:"cpuUsage"`
	ContainerID       string  `json:"containerId"`
	Percentage        float64 `json:"percentage"`
	ErrorMessage      string  `json:"errorMessage"`
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
	ApplicationID     int                         `json:"applicationID"`
	Application       string                      `json:"application"`
	CatalogItemID     int                         `json:"catalogItemId"`
	AgentUUID         string                      `json:"iofogUuid"`
	UserID            int                         `json:"userId"`
	RegistryID        int                         `json:"registryId"`
	Ports             []MicroservicePortMapping   `json:"ports"`
	Volumes           []MicroserviceVolumeMapping `json:"volumeMappings"`
	Commands          []string                    `json:"cmd"`
	Env               []MicroserviceEnvironment   `json:"env"`
	ExtraHosts        []MicroserviceExtraHost     `json:"extraHosts"`
	Status            MicroserviceStatus          `json:"status"`
	Images            []CatalogImage              `json:"images"`
}

type MicroserviceExtraHost struct {
	Name    string `json:"name,omitempty"`
	Address string `json:"address,omitempty"`
	Value   string `json:"value,omitempty"`
}

type MicroserviceCreateRequest struct {
	Config         string                      `json:"config"`
	Name           string                      `json:"name"`
	RootHostAccess interface{}                 `json:"rootHostAccess"`
	LogSize        int                         `json:"logSize"`
	FlowID         int                         `json:"flowId"`
	Application    string                      `json:"application"`
	CatalogItemID  int                         `json:"catalogItemId,omitempty"`
	AgentUUID      string                      `json:"iofogUuid,omitempty"`
	AgentName      string                      `json:"agentName,omitempty"`
	RegistryID     int                         `json:"registryId"`
	Ports          []MicroservicePortMapping   `json:"ports"`
	Volumes        []MicroserviceVolumeMapping `json:"volumeMappings"`
	Commands       []string                    `json:"cmd,omitempty"`
	Env            []MicroserviceEnvironment   `json:"env"`
	Images         []CatalogImage              `json:"images,omitempty"`
	ExtraHosts     []MicroserviceExtraHost     `json:"extraHosts,omitempty"`
}

type MicroserviceUpdateRequest struct {
	UUID              string                       `json:"-"`
	Config            *string                      `json:"config,omitempty"`
	Name              *string                      `json:"name,omitempty"`
	RootHostAccess    interface{}                  `json:"rootHostAccess,omitempty"`
	LogSize           *int                         `json:"logSize,omitempty"`
	Delete            *bool                        `json:"delete,omitempty"`
	DeleteWithCleanup *bool                        `json:"deleteWithCleanup,omitempty"`
	FlowID            *int                         `json:"flowId,omitempty"`
	Application       *string                      `json:"application,omitempty"`
	AgentUUID         *string                      `json:"iofogUuid,omitempty"`
	AgentName         *string                      `json:"agentName,omitempty"`
	UserID            *int                         `json:"userId,omitempty"`
	RegistryID        *int                         `json:"registryId,omitempty"`
	CatalogItemID     int                          `json:"catalogItemId,omitempty"`
	Ports             []MicroservicePortMapping    `json:"-"` // Ports are not valid in Controller PATCH call, need to use separate API calls
	Volumes           *[]MicroserviceVolumeMapping `json:"volumeMappings,omitempty"`
	Commands          *[]string                    `json:"cmd,omitempty"`
	Env               *[]MicroserviceEnvironment   `json:"env,omitempty"`
	ExtraHosts        *[]MicroserviceExtraHost     `json:"extraHosts,omitempty"`
	Images            []CatalogImage               `json:"images,omitempty"`
	Rebuild           interface{}                  `json:"rebuild,omitempty"`
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

type MicroservicePublicPort struct {
	MicroserviceUUID string     `json:"microserviceUuid"`
	PublicPort       PublicPort `json:"publicPort"`
}

type PublicPort struct {
	Protocol string `json:"protocol"`
	Queue    string `json:"queueName"`
	Port     int    `json:"publicPort"`
}

// Users

type User struct {
	Name     string `json:"firstName"`
	Surname  string `json:"lastName"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ControllerVersions struct {
	Controller string `json:"controller"`
	EcnViewer  string `json:"ecnViewer"`
}

type ControllerStatus struct {
	Status        string             `json:"status"`
	UptimeSeconds float64            `json:"uptimeSec"`
	Versions      ControllerVersions `json:"versions"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string `json:"accessToken"`
}

type ListAgentsRequest struct {
	System  bool              `json:"system"`
	Filters []AgentListFilter `json:"filters"`
}

type CreateAgentRequest struct {
	AgentUpdateRequest `json:",inline"`
}

type CreateAgentResponse struct {
	UUID string
}

type GetAgentProvisionKeyResponse struct {
	Key             string `json:"key"`
	ExpireTimeMsUTC int64  `json:"expirationTime"`
}

type AgentInfo struct {
	UUID                      string    `json:"uuid" yaml:"uuid"`
	Name                      string    `json:"name" yaml:"name"`
	Host                      string    `json:"host" yaml:"host"`
	Location                  string    `json:"location" yaml:"location"`
	Latitude                  float64   `json:"latitude" yaml:"latitude"`
	Longitude                 float64   `json:"longitude" yaml:"longitude"`
	Description               string    `json:"description" yaml:"description"`
	DockerURL                 string    `json:"dockerUrl" yaml:"dockerUrl"`
	DiskLimit                 int64     `json:"diskLimit" yaml:"diskLimit"`
	DiskDirectory             string    `json:"diskDirectory" yaml:"diskDirectory"`
	MemoryLimit               int64     `json:"memoryLimit" yaml:"memoryLimit"`
	CPULimit                  int64     `json:"cpuLimit" yaml:"cpuLimit"`
	LogLimit                  int64     `json:"logLimit" yaml:"logLimit"`
	LogDirectory              string    `json:"logDirectory" yaml:"logDirectory"`
	LogFileCount              int64     `json:"logFileCount" yaml:"logFileCount"`
	StatusFrequency           float64   `json:"statusFrequency" yaml:"statusFrequency"`
	ChangeFrequency           float64   `json:"changeFrequency" yaml:"changeFrequency"`
	DeviceScanFrequency       float64   `json:"deviceScanFrequency" yaml:"deviceScanFrequency"`
	BluetoothEnabled          bool      `json:"bluetoothEnabled" yaml:"bluetoothEnabled"`
	WatchdogEnabled           bool      `json:"watchdogEnabled" yaml:"watchdogEnabled"`
	AbstractedHardwareEnabled bool      `json:"abstractedHardwareEnabled" yaml:"abstractedHardwareEnabled"`
	CreatedTimeRFC3339        string    `json:"createdAt" yaml:"created"`
	UpdatedTimeRFC3339        string    `json:"updatedAt" yaml:"updated"`
	LastActive                int64     `json:"lastActive" yaml:"lastActive"`
	DaemonStatus              string    `json:"daemonStatus" yaml:"daemonStatus"`
	UptimeMs                  int64     `json:"daemonOperatingDuration" yaml:"uptime"`
	MemoryUsage               float64   `json:"memoryUsage" yaml:"memoryUsage"`
	DiskUsage                 float64   `json:"diskUsage" yaml:"diskUsage"`
	CPUUsage                  float64   `json:"cpuUsage" yaml:"cpuUsage"`
	MemoryViolation           string    `json:"memoryViolation" yaml:"memoryViolation"`
	DiskViolation             string    `json:"diskViolation" yaml:"diskViolation"`
	CPUViolation              string    `json:"cpuViolation" yaml:"cpuViolation"`
	MicroserviceStatus        string    `json:"microserviceStatus" yaml:"microserviceStatus"`
	RepositoryCount           int64     `json:"repositoryCount" yaml:"repositoryCount"`
	RepositoryStatus          string    `json:"repositoryStatus" yaml:"repositoryStatus"`
	LastStatusTimeMsUTC       int64     `json:"lastStatusTime" yaml:"lastStatusTime"`
	IPAddress                 string    `json:"ipAddress" yaml:"ipAddress"`
	IPAddressExternal         string    `json:"ipAddressExternal" yaml:"ipAddressExternal"`
	ProcessedMessaged         int64     `json:"processedMessages" yaml:"ProcessedMessages"`
	MicroserviceMessageCount  int64     `json:"microserviceMessageCounts" yaml:"microserviceMessageCount"`
	MessageSpeed              float64   `json:"messageSpeed" yaml:"messageSpeed"`
	LastCommandTimeMsUTC      int64     `json:"lastCommandTime" yaml:"lastCommandTime"`
	NetworkInterface          string    `json:"networkInterface" yaml:"networkInterface"`
	Version                   string    `json:"version" yaml:"version"`
	IsReadyToUpgrade          bool      `json:"isReadyToUpgrade" yaml:"isReadyToUpgrade"`
	IsReadyToRollback         bool      `json:"isReadyToRollback" yaml:"isReadyToRollback"`
	Tunnel                    string    `json:"tunnel" yaml:"tunnel"`
	FogType                   int       `json:"fogTypeId" yaml:"fogTypeId"`
	RouterMode                string    `json:"routerMode" yaml:"routerMode"`
	NetworkRouter             *string   `json:"networkRouter,omitempty" yaml:"networkRouter,omitempty"`
	UpstreamRouters           *[]string `json:"upstreamRouters,omitempty" yaml:"upstreamRouters,omitempty"`
	MessagingPort             *int      `json:"messagingPort,omitempty" yaml:"messagingPort,omitempty"`
	EdgeRouterPort            *int      `json:"edgeRouterPort,omitempty" yaml:"edgeRouterPort,omitempty"`
	InterRouterPort           *int      `json:"interRouterPort,omitempty" yaml:"interRouterPort,omitempty"`
	LogLevel                  *string   `json:"logLevel" yaml:"logLevel"`
	DockerPruningFrequency    *float64  `json:"dockerPruningFrequency" yaml:"dockerPruningFrequency"`
	AvailableDiskThreshold    *float64  `json:"availableDiskThreshold" yaml:"availableDiskThreshold"`
	Tags                      *[]string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

type RouterConfig struct {
	RouterMode      *string `json:"routerMode,omitempty" yaml:"routerMode,omitempty"`
	MessagingPort   *int    `json:"messagingPort,omitempty" yaml:"messagingPort,omitempty"`
	EdgeRouterPort  *int    `json:"edgeRouterPort,omitempty" yaml:"edgeRouterPort,omitempty"`
	InterRouterPort *int    `json:"interRouterPort,omitempty" yaml:"interRouterPort,omitempty"`
}

type AgentConfiguration struct {
	DockerURL                 *string   `json:"dockerUrl,omitempty" yaml:"dockerUrl"`
	DiskLimit                 *int64    `json:"diskLimit,omitempty" yaml:"diskLimit"`
	DiskDirectory             *string   `json:"diskDirectory,omitempty" yaml:"diskDirectory"`
	MemoryLimit               *int64    `json:"memoryLimit,omitempty" yaml:"memoryLimit"`
	CPULimit                  *int64    `json:"cpuLimit,omitempty" yaml:"cpuLimit"`
	LogLimit                  *int64    `json:"logLimit,omitempty" yaml:"logLimit"`
	LogDirectory              *string   `json:"logDirectory,omitempty" yaml:"logDirectory"`
	LogFileCount              *int64    `json:"logFileCount,omitempty" yaml:"logFileCount"`
	StatusFrequency           *float64  `json:"statusFrequency,omitempty" yaml:"statusFrequency"`
	ChangeFrequency           *float64  `json:"changeFrequency,omitempty" yaml:"changeFrequency"`
	DeviceScanFrequency       *float64  `json:"deviceScanFrequency,omitempty" yaml:"deviceScanFrequency"`
	BluetoothEnabled          *bool     `json:"bluetoothEnabled,omitempty" yaml:"bluetoothEnabled"`
	WatchdogEnabled           *bool     `json:"watchdogEnabled,omitempty" yaml:"watchdogEnabled"`
	AbstractedHardwareEnabled *bool     `json:"abstractedHardwareEnabled,omitempty" yaml:"abstractedHardwareEnabled"`
	IsSystem                  *bool     `json:"isSystem,omitempty" yaml:"-"` // Can't specify system agent using yaml file.
	UpstreamRouters           *[]string `json:"upstreamRouters,omitempty" yaml:"upstreamRouters,omitempty"`
	NetworkRouter             *string   `json:"networkRouter,omitempty" yaml:"networkRouter,omitempty"`
	Host                      *string   `json:"host,omitempty" yaml:"host,omitempty"`
	RouterConfig              `json:",omitempty" yaml:"routerConfig,omitempty"`
	LogLevel                  *string  `json:"logLevel,omitempty" yaml:"logLevel"`
	DockerPruningFrequency    *float64 `json:"dockerPruningFrequency,omitempty" yaml:"dockerPruningFrequency"`
	AvailableDiskThreshold    *float64 `json:"availableDiskThreshold,omitempty" yaml:"availableDiskThreshold"`
}

type AgentUpdateRequest struct {
	UUID        string    `json:"-"`
	Name        string    `json:"name,omitempty" yaml:"name"`
	Location    string    `json:"location,omitempty" yaml:"location"`
	Latitude    float64   `json:"latitude,omitempty" yaml:"latitude"`
	Longitude   float64   `json:"longitude,omitempty" yaml:"longitude"`
	Description string    `json:"description,omitempty" yaml:"description"`
	FogType     *int64    `json:"fogType,omitempty" yaml:"agentType"`
	Tags        *[]string `json:"tags,omitempty" yaml:"tags"`
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

type Router struct {
	RouterConfig
	Host string `json:"host"`
}

type UpdateConfigRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func newDefaultProxyRequest(address string) *UpdateConfigRequest {
	return &UpdateConfigRequest{
		Key:   "default-proxy-host",
		Value: address,
	}
}

func newPublicPortHostRequest(protocol Protocol, host string) *UpdateConfigRequest {
	return &UpdateConfigRequest{
		Key:   protocol + "-public-port-host",
		Value: host,
	}
}

type RouteListResponse struct {
	Routes []Route `json:"routes"`
}

type Route struct {
	Name                   string `json:"name"`
	Application            string `json:"application"`
	SourceMicroserviceUUID string `json:"sourceMicroserviceUuid"`
	DestMicroserviceUUID   string `json:"destMicroserviceUuid"`
}

type ApplicationRouteCreateRequest struct {
	Name string `json:"name"`
	From string `json:"from"`
	To   string `json:"to"`
}

type EdgeResourceDisplay struct {
	Name  string `json:"name,omitempty"`
	Icon  string `json:"icon,omitempty"`
	Color string `json:"color,omitempty"`
}

type EdgeResourceMetadata struct {
	Name              string               `json:"name,omitempty"`
	Description       string               `json:"description,omitempty"`
	Version           string               `json:"version,omitempty"`
	InterfaceProtocol string               `json:"interfaceProtocol,omitempty"`
	Display           *EdgeResourceDisplay `json:"display,omitempty"`
	Interface         HTTPEdgeResource     `json:"interface,omitempty"` // TODO: Make this generic
	OrchestrationTags []string             `json:"orchestrationTags,omitempty"`
}

type HTTPEdgeResource struct {
	Endpoints []HTTPEndpoint `json:"endpoints,omitempty"`
}

type HTTPEndpoint struct {
	Name   string `json:"name,omitempty"`
	Method string `json:"method,omitempty"`
	URL    string `json:"url,omitempty"`
}

type LinkEdgeResourceRequest struct {
	AgentUUID           string `json:"uuid"`
	EdgeResourceName    string `json:"-"`
	EdgeResourceVersion string `json:"-"`
}

type ListEdgeResourceResponse struct {
	EdgeResources []EdgeResourceMetadata `json:"edgeResources"`
}
