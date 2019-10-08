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

package deploy

import "github.com/eclipse-iofog/iofog-go-sdk/pkg/client"

// HeaderMetadata contains k8s metadata
type HeaderMetadata struct {
	Name      string
	Namespace string
}

// Deployable can be deployed using iofogctl deploy
type Deployable interface {
	Deploy(namespace string) error
}

// Kind contains available types
type Kind string

// Available kind of deploy
const (
	ApplicationKind  Kind = "iofog-application"
	MicroserviceKind Kind = "iofog-microservice"
	ControlPlaneKind Kind = "iofog-controlplane"
	AgentKind        Kind = "iofog-agent"
	ConnectorKind    Kind = "iofog-connector"
)

// Header contains k8s yaml header
type Header struct {
	APIVersion string         `yaml:"apiVersion"`
	Kind       Kind           `yaml:"kind"`
	Metadata   HeaderMetadata `yaml:"metadata"`
	Spec       interface{}
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

// Microservices is a list of Microservice
type Microservices struct {
	Microservices []Microservices
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

// Applications is a list of applications
type Applications struct {
	Applications []Application
}

// IofogController contains informations needed to connect to the controller
type IofogController struct {
	Email    string
	Password string
	Endpoint string
}
