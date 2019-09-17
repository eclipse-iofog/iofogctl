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

package config

import "github.com/eclipse-iofog/iofog-go-sdk/pkg/client"

// IofogUser contains information about users registered against a controller
type IofogUser struct {
	Name     string
	Surname  string
	Email    string
	Password string
}

// DockerCredentials credentials used to log into docker when deploying a local stack
type DockerCredentials struct {
	User     string
	Password string
}

type Database struct {
	Provider     string
	Host         string
	Port         int
	User         string
	Password     string
	DatabaseName string
}

type Loadbalancer struct {
	Host string
	Port int
}

type ControlPlane struct {
	Database     Database
	LoadBalancer Loadbalancer
	IofogUser    IofogUser
	Controllers  []Controller
	Images       map[string]string
}

type Connector struct {
	Name              string
	User              string
	Host              string
	Port              int
	KeyFile           string
	KubeConfig        string
	KubeConnectorIP   string
	Image             string
	ImageCredentials  DockerCredentials // Optional credentials if needed to pull image
	Created           string
	Version           string
	Endpoint          string
	PackageCloudToken string
}

// Controller contains information for configuring a controller
type Controller struct {
	Name              string
	User              string
	Host              string
	Port              int
	KeyFile           string
	KubeConfig        string
	Replicas          int
	KubeControllerIP  string
	Endpoint          string
	Created           string
	ImageCredentials  DockerCredentials // Optional credentials if needed to pull images
	Version           string
	PackageCloudToken string
}

// Agent contains information for configuring an agent
type Agent struct {
	Name              string
	User              string
	Host              string
	Port              int
	KeyFile           string
	UUID              string
	Created           string
	Image             string
	ImageCredentials  DockerCredentials // Optional credentials if needed to pull image
	Version           string
	PackageCloudToken string
}

// CatalogItem contains information about a catalog item
type CatalogItem struct {
	ID            int
	X86           string
	ARM           string
	Registry      string
	Name          string
	Description   string
	ConfigExample string
}

// MicroserviceImages contains information about the images for a microservice
type MicroserviceImages struct {
	CatalogID int
	X86       string
	ARM       string
	Registry  string
}

// MicroserviceAgent contains information about required agent configuration for a microservice
type MicroserviceAgent struct {
	Name   string
	Config client.AgentConfiguration
}

// Microservice contains information for configuring a microservice
type Microservice struct {
	UUID           string `yaml:"-"`
	Name           string
	Agent          MicroserviceAgent
	Images         MicroserviceImages
	Config         map[string]interface{}
	RootHostAccess bool
	Ports          []client.MicroservicePortMapping   `yaml:"ports"`
	Volumes        []client.MicroserviceVolumeMapping `yaml:"volumes"`
	Env            []client.MicroserviceEnvironment   `yaml:"env"`
	Routes         []string                           `yaml:"routes,omitempty"`
	Flow           *string                            `yaml:"application,omitempty"`
	Created        string                             `yaml:"created,omitempty"`
}

// Route contains information about a route from one microservice to another
type Route struct {
	From string
	To   string
}

// Application contains information for configuring an application
type Application struct {
	Name          string
	Microservices []Microservice
	Routes        []Route
	ID            int
}

// Namespace contains information for configuring a namespace
type Namespace struct {
	Name          string
	ControlPlane  ControlPlane
	Connectors    []Connector
	Agents        []Agent
	Microservices []Microservice
	Created       string
}

// configuration contains the unmarshalled configuration file
type configuration struct {
	Namespaces []Namespace
}
