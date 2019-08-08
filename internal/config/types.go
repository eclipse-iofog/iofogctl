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

import "github.com/eclipse-iofog/iofogctl/pkg/iofog/client"

// IofogUser contains information about users registered against a controller
type IofogUser struct {
	Name     string `mapstructure:"name"`
	Surname  string `mapstructure:"surname"`
	Email    string `mapstructure:"email"`
	Password string `mapstructure:"password"`
}

type Database struct {
	Type     string
	Host     string
	Port     int
	User     string
	Password string
}

type Loadbalancer struct {
	Host string
	Port int
}

type ControlPlane struct {
	Database     Database
	LoadBalancer Loadbalancer
	Controllers  []Controller
}

type Connector struct {
	// TODO
}

// Controller contains information for configuring a controller
type Controller struct {
	Name              string            `mapstructure:"name"`
	User              string            `mapstructure:"user"`
	Host              string            `mapstructure:"host"`
	Port              int               `mapstructure:"port"`
	KeyFile           string            `mapstructure:"keyfile"`
	KubeConfig        string            `mapstructure:"kubeconfig"`
	KubeControllerIP  string            `mapstructure:"kubecontrollerip"`
	Endpoint          string            `mapstructure:"endpoint"`
	IofogUser         IofogUser         `mapstructure:"iofoguser"`
	Created           string            `mapstructure:"created"`
	Images            map[string]string `mapstructure:"images"`
	Version           string            `mapstructure:"version"`
	PackageCloudToken string            `mapstructure:"packagecloudtoken"`
}

// Agent contains information for configuring an agent
type Agent struct {
	Name    string `mapstructure:"name"`
	User    string `mapstructure:"user"`
	Host    string `mapstructure:"host"`
	Port    int    `mapstructure:"port"`
	KeyFile string `mapstructure:"keyfile"`
	UUID    string `mapstructure:"uuid"`
	Created string `mapstructure:"created"`
	Image   string `mapstructure:"image"`
}

// MicroserviceImages contains information about the images for a microservice
type MicroserviceImages struct {
	CatalogID int
	X86       string
	ARM       string
	Registry  int
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
	Ports          []client.MicroservicePortMapping   `yaml:"ports,omitempty"`
	Volumes        []client.MicroserviceVolumeMapping `yaml:"volumes,omitempty"`
	Env            []client.MicroserviceEnvironment   `yaml:"env,omitempty"`
	Routes         []string                           `yaml:"routes,omitempty"`
	Flow           *int                               `yaml:"flow,omitempty"`
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
}

// Namespace contains information for configuring a namespace
type Namespace struct {
	Name          string `mapstructure:"name"`
	ControlPlane  ControlPlane
	Connectors    []Connector
	Agents        []Agent        `mapstructure:"agents"`
	Microservices []Microservice `mapstructure:"microservices"`
	Created       string         `mapstructure:"created"`
}

// configuration contains the unmarshalled configuration file
type configuration struct {
	Namespaces []Namespace `mapstructure:"namespaces"`
}
