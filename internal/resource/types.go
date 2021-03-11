/*
 *  *******************************************************************************
 *  * Copyright (c) 2020 Edgeworx, Inc.
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package resource

import (
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/apps"
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
)

type Route = apps.Route
type Microservice = apps.Microservice
type Application = apps.Application
type ApplicationTemplate = apps.ApplicationTemplate

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
	Controller  string `yaml:"controller,omitempty"`
	Operator    string `yaml:"operator,omitempty"`
	Kubelet     string `yaml:"kubelet,omitempty"`
	PortManager string `yaml:"portManager,omitempty"`
	Router      string `yaml:"router,omitempty"`
	Proxy       string `yaml:"proxy,omitempty"`
}

type Services struct {
	Controller Service `json:"controller,omitempty"`
	Router     Service `json:"router,omitempty"`
	Proxy      Service `json:"proxy,omitempty"`
}

type Service struct {
	Type string `json:"type,omitempty"`
	IP   string `json:"ip,omitempty"`
}

type Replicas struct {
	Controller int32 `yaml:"controller"`
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

type Registry struct {
	URL          *string `yaml:"url"`
	Private      *bool   `yaml:"private"`
	Username     *string `yaml:"username"`
	Password     *string `yaml:"password"`
	Email        *string `yaml:"email"`
	RequiresCert *bool   `yaml:"requiresCert"`
	Certificate  *string `yaml:"certificate,omitempty"`
	ID           int     `yaml:"id"`
}

type Volume struct {
	Name        string   `json:"name" yaml:"name"`
	Agents      []string `json:"agents" yaml:"agents"`
	Source      string   `json:"source" yaml:"source"`
	Destination string   `json:"destination" yaml:"destination"`
	Permissions string   `json:"permissions" yaml:"permissions"`
}

// AgentConfiguration contains configuration information for a deployed agent
type AgentConfiguration struct {
	Name                      string  `json:"name,omitempty" yaml:"name"`
	Location                  string  `json:"location,omitempty" yaml:"location"`
	Latitude                  float64 `json:"latitude,omitempty" yaml:"latitude"`
	Longitude                 float64 `json:"longitude,omitempty" yaml:"longitude"`
	Description               string  `json:"description,omitempty" yaml:"description"`
	FogType                   *string `json:"fogType,omitempty" yaml:"agentType"`
	client.AgentConfiguration `yaml:",inline"`
}

type EdgeResource struct {
	Name              string
	Version           string                     `yaml:"version"`
	Description       string                     `yaml:"description"`
	InterfaceProtocol string                     `yaml:"interfaceProtocol"`
	Interface         *EdgeResourceHTTPInterface `yaml:"interface,omitempty"` // TODO: Make this generic to support multiple interfaces protocols
	Display           *Display                   `yaml:"display,omitempty"`
	OrchestrationTags []string                   `yaml:"orchestrationTags"`
	Custom            map[string]interface{}     `yaml:"custom"`
}

type EdgeResourceHTTPInterface = client.HTTPEdgeResource

type Display = client.EdgeResourceDisplay
type HTTPEndpoint = client.HTTPEndpoint

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

type ControllerConfig struct {
	PidBaseDir    string `yaml:"pidBaseDir,omitempty"`
	EcnViewerPort int    `yaml:"ecnViewerPort,omitempty"`
}
