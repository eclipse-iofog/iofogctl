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

import "github.com/eclipse-iofog/iofog-go-sdk/pkg/apps"

// IofogUser contains information about users registered against a controller
type IofogUser struct {
	Name     string `yaml:"name,omitempty"`
	Surname  string `yaml:"surname,omitempty"`
	Email    string `yaml:"email,omitempty"`
	Password string `yaml:"password,omitempty"`
}

// DockerCredentials credentials used to log into docker when deploying a local stack
type DockerCredentials struct {
	User     string `yaml:"user,omitempty"`
	Password string `yaml:"password,omitempty"`
}

type Database struct {
	Provider     string `yaml:"provider,omitempty"`
	Host         string `yaml:"host,omitempty"`
	Port         int    `yaml:"port,omitempty"`
	User         string `yaml:"user,omitempty"`
	Password     string `yaml:"password,omitempty"`
	DatabaseName string `yaml:"databaseName,omitempty"`
}

type Loadbalancer struct {
	Host string `yaml:"host,omitempty"`
	Port int    `yaml:"port,omitempty"`
}

type ControlPlane struct {
	Database     Database          `yaml:"database,omitempty"`
	LoadBalancer Loadbalancer      `yaml:"loadBalancer,omitempty"`
	IofogUser    IofogUser         `yaml:"iofogUser,omitempty"`
	Controllers  []Controller      `yaml:"controllers,omitempty"`
	Images       map[string]string `yaml:"images,omitempty"`
}

type Connector struct {
	Name              string            `yaml:"name,omitempty"`
	User              string            `yaml:"user,omitempty"`
	Host              string            `yaml:"host,omitempty"`
	Port              int               `yaml:"port,omitempty"`
	KeyFile           string            `yaml:"keyFile,omitempty"`
	KubeConfig        string            `yaml:"kubeConfig,omitempty"`
	KubeConnectorIP   string            `yaml:"kubeConnectorIP,omitempty"`
	Image             string            `yaml:"image,omitempty"`
	ImageCredentials  DockerCredentials `yaml:"imageCredentials,omitempty"` // Optional credentials if needed to pull image
	Created           string            `yaml:"created,omitempty"`
	Version           string            `yaml:"version,omitempty"`
	Endpoint          string            `yaml:"endpoint,omitempty"`
	PackageCloudToken string            `yaml:"packageCloudToken,omitempty"`
}

// Controller contains information for configuring a controller
type Controller struct {
	Name              string            `yaml:"name,omitempty"`
	User              string            `yaml:"user,omitempty"`
	Host              string            `yaml:"host,omitempty"`
	Port              int               `yaml:"port,omitempty"`
	KeyFile           string            `yaml:"keyFile,omitempty"`
	KubeConfig        string            `yaml:"kubeConfig,omitempty"`
	Replicas          int               `yaml:"replicas,omitempty"`
	ServiceType       string            `yaml:"serviceType,omitempty"`
	KubeControllerIP  string            `yaml:"kubeControllerIP,omitempty"`
	Endpoint          string            `yaml:"endpoint,omitempty"`
	Created           string            `yaml:"created,omitempty"`
	ImageCredentials  DockerCredentials `yaml:"imageCredentials,omitempty"` // Optional credentials if needed to pull images
	Version           string            `yaml:"version,omitempty"`
	PackageCloudToken string            `yaml:"packageCloudToken,omitempty"`
	Repo              string            `yaml:",omitempty"`
	Token             string            `yaml:",omitempty"`
}

// Agent contains information for configuring an agent
type Agent struct {
	Name              string            `yaml:"name,omitempty"`
	User              string            `yaml:"user,omitempty"`
	Host              string            `yaml:"host,omitempty"`
	Port              int               `yaml:"port,omitempty"`
	KeyFile           string            `yaml:"keyFile,omitempty"`
	UUID              string            `yaml:"uuid,omitempty"`
	Created           string            `yaml:"created,omitempty"`
	Image             string            `yaml:"image,omitempty"`
	ImageCredentials  DockerCredentials `yaml:"imageCredentials,omitempty"` // Optional credentials if needed to pull image
	Version           string            `yaml:"version,omitempty"`
	PackageCloudToken string            `yaml:"packageCloudToken,omitempty"`
	Repo              string            `yaml:",omitempty"`
	Token             string            `yaml:",omitempty"`
}

// Namespace contains information for configuring a namespace
type Namespace struct {
	Name         string       `yaml:"name,omitempty"`
	ControlPlane ControlPlane `yaml:"controlPlane,omitempty"`
	Connectors   []Connector  `yaml:"connectors,omitempty"`
	Agents       []Agent      `yaml:"agents,omitempty"`
	Created      string       `yaml:"created,omitempty"`
}

// configuration contains the unmarshalled configuration file
type configuration struct {
	Namespaces []Namespace `yaml:"namespaces,omitempty"`
}

// HeaderMetadata contains k8s metadata
type HeaderMetadata struct {
	Name      string `yaml:"name" json:"name"`
	Namespace string `yaml:"namespace" json:"namespace"`
}

// Header contains k8s yaml header
type Header struct {
	APIVersion string         `yaml:"apiVersion" json:"apiVersion"`
	Kind       apps.Kind      `yaml:"kind" json:"kind"`
	Metadata   HeaderMetadata `yaml:"metadata" json:"metadata"`
	Spec       interface{}    `yaml:"spec" json:"spec"`
}
