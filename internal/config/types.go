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

import (
	"github.com/eclipse-iofog/iofog-go-sdk/pkg/apps"
	"github.com/eclipse-iofog/iofog-go-sdk/pkg/client"
)

// iofogctl specific Kinds
const (
	AgentConfigKind       apps.Kind = "AgentConfig"
	CatalogItemKind       apps.Kind = "CatalogItem"
	IofogctlConfigKind    apps.Kind = "IofogctlConfig"
	IofogctlNamespaceKind apps.Kind = "IofogctlNamespace"
)

type Container struct {
	Image       string      `yaml:"image,omitempty"`
	Credentials Credentials `yaml:"credentials,omitempty"` // Optional credentials if needed to pull images
}

type Package struct {
	Version string `yaml:"version,omitempty"`
	Repo    string `yaml:"repo,omitempty"`
	Token   string `yaml:"token,omitempty"`
}

type SSH struct {
	User    string `yaml:"user,omitempty"`
	Port    int    `yaml:"port,omitempty"`
	KeyFile string `yaml:"keyFile,omitempty"`
}

type KubeImages struct {
	Operator string `yaml:"operator,omitempty"`
	Kubelet  string `yaml:"kubelet,omitempty"`
}

type Kube struct {
	Config      string     `yaml:"config,omitempty"`
	StaticIP    string     `yaml:"staticIp,omitempty"`
	Replicas    int        `yaml:"replicas,omitempty"`
	ServiceType string     `yaml:"serviceType,omitempty"`
	Images      KubeImages `yaml:"images,omitempty"`
}

// IofogUser contains information about users registered against a controller
type IofogUser struct {
	Name     string `yaml:"name,omitempty"`
	Surname  string `yaml:"surname,omitempty"`
	Email    string `yaml:"email,omitempty"`
	Password string `yaml:"password,omitempty"`
}

// Credentials credentials used to log into docker when deploying a local stack
type Credentials struct {
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
	Database     Database     `yaml:"database,omitempty"`
	LoadBalancer Loadbalancer `yaml:"loadBalancer,omitempty"`
	IofogUser    IofogUser    `yaml:"iofogUser,omitempty"`
	Controllers  []Controller `yaml:"controllers,omitempty"`
}

type Connector struct {
	Name      string    `yaml:"name,omitempty"`
	Host      string    `yaml:"host,omitempty"`
	SSH       SSH       `yaml:"ssh,omitempty"`
	Kube      Kube      `yaml:"kube,omitempty"`
	Created   string    `yaml:"created,omitempty"`
	Endpoint  string    `yaml:"endpoint,omitempty"`
	Package   Package   `yaml:"package,omitempty"`
	Container Container `yaml:"container,omitempty"`
}

// Controller contains information for configuring a controller
type Controller struct {
	Name      string    `yaml:"name,omitempty"`
	Host      string    `yaml:"host,omitempty"`
	SSH       SSH       `yaml:"ssh,omitempty"`
	Kube      Kube      `yaml:"kube,omitempty"`
	Endpoint  string    `yaml:"endpoint,omitempty"`
	Created   string    `yaml:"created,omitempty"`
	Package   Package   `yaml:"package,omitempty"`
	Container Container `yaml:"container,omitempty"`
}

// AgentConfiguration contains configuration information for a deployed agent
type AgentConfiguration struct {
	Name                      string  `json:"name,omitempty" yaml:"name"`
	Location                  string  `json:"location,omitempty" yaml:"location"`
	Latitude                  float64 `json:"latitude,omitempty" yaml:"latitude"`
	Longitude                 float64 `json:"longitude,omitempty" yaml:"longitude"`
	Description               string  `json:"description,omitempty" yaml:"description"`
	FogType                   string  `json:"fogType,omitempty" yaml:"agentType"`
	client.AgentConfiguration `yaml:",inline"`
}

// FogTypeStringMap map human readable fog type to Controller fog type
var FogTypeStringMap = map[string]int64{
	"auto": 0,
	"x86":  1,
	"arm":  2,
}

// FogTypeIntMap map Controller fog type to human readable fog type
var FogTypeIntMap = map[int]string{
	0: "auto",
	1: "x86",
	2: "arm",
}

// Agent contains information for deploying an agent
type Agent struct {
	Name      string    `yaml:"name,omitempty"`
	Host      string    `yaml:"host,omitempty"`
	SSH       SSH       `yaml:"ssh,omitempty"`
	UUID      string    `yaml:"uuid,omitempty"`
	Created   string    `yaml:"created,omitempty"`
	Container Container `yaml:"container,omitempty"`
	Package   Package   `yaml:"package,omitempty"`
}

// Namespace contains information for configuring a namespace
type Namespace struct {
	Name         string       `yaml:"name,omitempty"`
	ControlPlane ControlPlane `yaml:"controlPlane,omitempty"`
	Connectors   []Connector  `yaml:"connectors,omitempty"`
	Agents       []Agent      `yaml:"agents,omitempty"`
	Created      string       `yaml:"created,omitempty"`
}

// Configuration contains the unmarshalled configuration file
type configuration struct {
	DefaultNamespace string   `yaml:"defaultNamespace"`
	CurrentNamespace string   `yaml:"-"`
	Namespaces       []string `yaml:"namespaces,omitempty"`
}

type iofogctlConfig struct {
	Header
}

type iofogctlNamespace struct {
	Header
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
