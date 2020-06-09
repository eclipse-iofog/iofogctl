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

package util

import "fmt"

// Set by linker
var (
	versionNumber = "undefined"
	platform      = "undefined"
	commit        = "undefined"
	date          = "undefined"

	repo = "undefined"

	controllerTag  = "undefined"
	agentTag       = "undefined"
	operatorTag    = "undefined"
	kubeletTag     = "undefined"
	routerTag      = "undefined"
	portManagerTag = "undefined"
	proxyTag       = "undefined"

	controllerVersion = "undefined"
	agentVersion      = "undefined"
)

const (
	controllerImage  = "controller"
	agentImage       = "agent"
	operatorImage    = "operator"
	kubeletImage     = "kubelet"
	portManagerImage = "port-manager"
	proxyImage       = "proxy"
	proxyARMImage    = "proxy-arm"
	routerImage      = "router"
	routerARMImage   = "router-arm"
)

type Version struct {
	VersionNumber string `yaml:"version"`
	Platform      string
	Commit        string
	Date          string
}

func GetVersion() Version {
	return Version{
		VersionNumber: versionNumber,
		Platform:      platform,
		Commit:        commit,
		Date:          date,
	}
}

func GetControllerVersion() string { return controllerVersion }
func GetAgentVersion() string      { return agentVersion }

func GetControllerImage() string {
	return fmt.Sprintf("%s/%s:%s", repo, controllerImage, controllerTag)
}
func GetAgentImage() string     { return fmt.Sprintf("%s/%s:%s", repo, agentImage, agentTag) }
func GetOperatorImage() string  { return fmt.Sprintf("%s/%s:%s", repo, operatorImage, operatorTag) }
func GetKubeletImage() string   { return fmt.Sprintf("%s/%s:%s", repo, kubeletImage, kubeletTag) }
func GetRouterImage() string    { return fmt.Sprintf("%s/%s:%s", repo, routerImage, routerTag) }
func GetRouterARMImage() string { return fmt.Sprintf("%s/%s:%s", repo, routerARMImage, routerTag) }
func GetPortManagerImage() string {
	return fmt.Sprintf("%s/%s:%s", repo, portManagerImage, portManagerTag)
}
func GetProxyImage() string    { return fmt.Sprintf("%s/%s:%s", repo, proxyImage, proxyTag) }
func GetProxyARMImage() string { return fmt.Sprintf("%s/%s:%s", repo, proxyARMImage, proxyTag) }
