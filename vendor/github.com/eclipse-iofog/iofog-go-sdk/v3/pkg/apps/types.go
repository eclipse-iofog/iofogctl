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

package apps

// HeaderMetadata contains k8s metadata
// +k8s:deepcopy-gen=true
type HeaderMetadata struct {
	Name      string `yaml:"name" json:"name"`
	Namespace string `yaml:"namespace" json:"namespace"`
}

// Kind contains available types
type Kind string

// IofogHeader represent the file structure
type IofogHeader Header

// Available kind of deploy
const (
	ApplicationKind         Kind = "Application"
	ApplicationTemplateKind Kind = "ApplicationTemplate"
	MicroserviceKind        Kind = "Microservice"
	RouteKind               Kind = "Route"
)

// Header contains k8s yaml header
type Header struct {
	APIVersion string         `yaml:"apiVersion" json:"apiVersion"`
	Kind       Kind           `yaml:"kind" json:"kind"`
	Metadata   HeaderMetadata `yaml:"metadata" json:"metadata"`
	Spec       interface{}    `yaml:"spec" json:"spec"`
}

// CatalogItem contains information about a catalog item
// +k8s:deepcopy-gen=true
type CatalogItem struct {
	ID            int    `yaml:"id" json:"id"`
	X86           string `yaml:"x86" json:"x86"`
	ARM           string `yaml:"arm" json:"arm"`
	Registry      string `yaml:"registry" json:"registry"`
	Name          string `yaml:"name" json:"name"`
	Description   string `yaml:"description" json:"description"`
	ConfigExample string `yaml:"configExample" json:"configExample"`
}

// MicroserviceImages contains information about the images for a microservice
// +k8s:deepcopy-gen=true
type MicroserviceImages struct {
	CatalogID int    `yaml:"catalogId" json:"catalogId"`
	X86       string `yaml:"x86" json:"x86"`
	ARM       string `yaml:"arm" json:"arm"`
	Registry  string `yaml:"registry" json:"registry"`
}

// MicroserviceAgent contains information about required agent configuration for a microservice
// +k8s:deepcopy-gen=true
type MicroserviceAgent struct {
	Name   string             `yaml:"name" json:"name"`
	Config AgentConfiguration `yaml:"config" json:"config"`
}

// MicroserviceContainer contains information for configuring a microservice container
// +k8s:deepcopy-gen=true
type MicroserviceContainer struct {
	Commands       []string                     `yaml:"commands,omitempty" json:"commands,omitempty"`
	Volumes        *[]MicroserviceVolumeMapping `yaml:"volumes,omitempty" json:"volumes,omitempty"`
	Env            *[]MicroserviceEnvironment   `yaml:"env,omitempty" json:"env,omitempty"`
	ExtraHosts     *[]MicroserviceExtraHost     `yaml:"extraHosts,omitempty" json:"extraHosts,omitempty"`
	Ports          []MicroservicePortMapping    `yaml:"ports" json:"ports"`
	RootHostAccess interface{}                  `yaml:"rootHostAccess" json:"rootHostAccess"` // +k8s:deepcopy-gen=ignore
}

// Microservice contains information for configuring a microservice
// +k8s:deepcopy-gen=true
type Microservice struct {
	UUID        string                `yaml:"uuid" json:"uuid"`
	Name        string                `yaml:"name" json:"name"`
	Agent       MicroserviceAgent     `yaml:"agent" json:"agent"`
	Images      *MicroserviceImages   `yaml:"images,omitempty" json:"images,omitempty"`
	Container   MicroserviceContainer `yaml:"container,omitempty" json:"container,omitempty"`
	Config      NestedMap             `yaml:"config" json:"config"`
	Flow        *string               `yaml:"flow,omitempty" json:"flow,omitempty"`
	Application *string               `yaml:"application,omitempty" json:"application,omitempty"`
	Created     string                `yaml:"created,omitempty" json:"created,omitempty"`
	Rebuild     interface{}           `yaml:"rebuild,omitempty" json:"rebuild,omitempty"` // +k8s:deepcopy-gen=ignore
}

type NestedMap map[string]interface{}

func (j NestedMap) DeepCopy() NestedMap {
	newMap := make(NestedMap)
	deepCopyNestedMap(j, newMap)
	return newMap
}

func deepCopyNestedMap(src, dest NestedMap) {
	for key := range src {
		switch value := src[key].(type) {
		case NestedMap:
			dest[key] = NestedMap{}
			deepCopyNestedMap(value, dest[key].(NestedMap))
		default:
			dest[key] = value
		}
	}
}

// +k8s:deepcopy-gen=true
type MicroservicePublicPortRouterInfo struct {
	Port int64  `json:"port"`
	Host string `json:"host"`
}

// +k8s:deepcopy-gen=true
type MicroservicePublicPortInfo struct {
	Schemes  []string                          `json:"schemes"`
	Links    []string                          `json:"links"`
	Protocol string                            `json:"protocol"`
	Enabled  bool                              `json:"enabled"`
	Router   *MicroservicePublicPortRouterInfo `yaml:"router,omitempty" json:"router,omitempty"`
}

// +k8s:deepcopy-gen=true
type MicroservicePortMapping struct {
	Internal int64                       `json:"internal"`
	External int64                       `json:"external"`
	Public   *MicroservicePublicPortInfo `yaml:"public,omitempty" json:"public,omitempty"`
	Protocol string                      `json:"protocol,omitempty"`
}

// +k8s:deepcopy-gen=true
type MicroserviceVolumeMapping struct {
	HostDestination      string `yaml:"hostDestination" json:"hostDestination"`
	ContainerDestination string `yaml:"containerDestination" json:"containerDestination"`
	AccessMode           string `yaml:"accessMode" json:"accessMode"`
	Type                 string `yaml:"type,omitempty" json:"type,omitempty"`
}

// +k8s:deepcopy-gen=true
type MicroserviceEnvironment struct {
	Key   string `yaml:"key" json:"key"`
	Value string `yaml:"value" json:"value"`
}

// +k8s:deepcopy-gen=true
type MicroserviceExtraHost struct {
	Name    string `yaml:"name" json:"name,omitempty"`
	Address string `yaml:"address" json:"address,omitempty"`
	Value   string `yaml:"value" json:"value,omitempty"`
}

// +k8s:deepcopy-gen=true
type AgentConfiguration struct {
	DockerURL                 *string   `yaml:"dockerUrl,omitempty" json:"dockerUrl,omitempty"`
	DiskLimit                 *int64    `yaml:"diskLimit,omitempty" json:"diskLimit,omitempty"`
	DiskDirectory             *string   `yaml:"diskDirectory,omitempty" json:"diskDirectory,omitempty"`
	MemoryLimit               *int64    `yaml:"memoryLimit,omitempty" json:"memoryLimit,omitempty"`
	CPULimit                  *int64    `yaml:"cpuLimit,omitempty" json:"cpuLimit,omitempty"`
	LogLimit                  *int64    `yaml:"logLimit,omitempty" json:"logLimit,omitempty"`
	LogDirectory              *string   `yaml:"logDirectory,omitempty" json:"logDirectory,omitempty"`
	LogFileCount              *int64    `yaml:"logFileCount,omitempty" json:"logFileCount,omitempty"`
	StatusFrequency           *float64  `yaml:"statusFrequency,omitempty" json:"statusFrequency,omitempty"`
	ChangeFrequency           *float64  `yaml:"changeFrequency,omitempty" json:"changeFrequency,omitempty"`
	DeviceScanFrequency       *float64  `yaml:"deviceScanFrequency,omitempty" json:"deviceScanFrequency,omitempty"`
	BluetoothEnabled          *bool     `yaml:"bluetoothEnabled,omitempty" json:"bluetoothEnabled,omitempty"`
	WatchdogEnabled           *bool     `yaml:"watchdogEnabled,omitempty" json:"watchdogEnabled,omitempty"`
	AbstractedHardwareEnabled *bool     `yaml:"abstractedHardwareEnabled,omitempty" json:"abstractedHardwareEnabled,omitempty"`
	RouterMode                *string   `yaml:"routerMode,omitempty" json:"routerMode,omitempty"`           // [edge, interior, none], default: edge
	RouterPort                *int      `yaml:"routerPort,omitempty" json:"routerPort,omitempty"`           // default: 5672
	UpstreamRouters           *[]string `yaml:"upstreamRouters,omitempty" json:"upstreamRouters,omitempty"` // ignored if routerMode: none
	NetworkRouter             *string   `yaml:"networkRouter,omitempty" json:"networkRouter,omitempty"`     // required if routerMone: none
}

// Microservices is a list of Microservice
// +k8s:deepcopy-gen=true
type Microservices struct {
	Microservices []Microservice `yaml:"microservices" json:"microservices"`
}

// Route contains information about a route from one microservice to another
// +k8s:deepcopy-gen=true
type Route struct {
	Name string `yaml:"name" json:"name"`
	From string `yaml:"from" json:"from"`
	To   string `yaml:"to" json:"to"`
}

// Application contains information for configuring an application
// +k8s:deepcopy-gen=true
type Application struct {
	Name          string               `yaml:"name" json:"name"`
	Microservices []Microservice       `yaml:"microservices,omitempty" json:"microservices,omitempty"`
	Routes        []Route              `yaml:"routes,omitempty" json:"routes,omitempty"`
	ID            int                  `yaml:"id,omitempty" json:"id,omitempty"`
	Template      *ApplicationTemplate `yaml:"template,omitempty" json:"template,omitempty"`
}

// ApplicationTemplate contains information for configuring an application template
// +k8s:deepcopy-gen=true
type ApplicationTemplate struct {
	Name        string                   `yaml:"name,omitempty"`
	Description string                   `yaml:"description,omitempty"`
	Variables   []TemplateVariable       `yaml:"variables,omitempty"`
	Application *ApplicationTemplateInfo `yaml:"application,omitempty"`
}

// TemplateVariable contains a key-value pair
// +k8s:deepcopy-gen=true
type TemplateVariable struct {
	Key          string      `yaml:"key"`
	Description  string      `yaml:"description"`
	DefaultValue interface{} `yaml:"defaultValue,omitempty"` // +k8s:deepcopy-gen=ignore
	Value        interface{} `yaml:"value,omitempty"`        // +k8s:deepcopy-gen=ignore
}

// ApplicationTemplateInfo contains microservice and route details for template
// +k8s:deepcopy-gen=true
type ApplicationTemplateInfo struct {
	Microservices []Microservice `yaml:"microservices"`
	Routes        []Route        `yaml:"routes"`
}

// Applications is a list of applications
// +k8s:deepcopy-gen=true
type Applications struct {
	Applications []Application `yaml:"applications" json:"applications"`
}

// IofogController contains informations needed to connect to the controller
// +k8s:deepcopy-gen=true
type IofogController struct {
	Email    string `yaml:"email" json:"email"`
	Password string `yaml:"password" json:"password"`
	Endpoint string `yaml:"endpoint" json:"endpoint"`
	Token    string `yaml:"token" json:"token"`
}
