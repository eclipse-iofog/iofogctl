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

package util

// Set by linker
var (
	versionNumber = "undefined"
	platform      = "undefined"
	commit        = "undefined"
	date          = "undefined"

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

func GetControllerTag() string  { return controllerTag }
func GetAgentTag() string       { return agentTag }
func GetOperatorTag() string    { return operatorTag }
func GetKubeletTag() string     { return kubeletTag }
func GetRouterTag() string      { return routerTag }
func GetPortManagerTag() string { return portManagerTag }
func GetProxyTag() string       { return proxyTag }

func GetControllerVersion() string { return controllerVersion }
func GetAgentVersion() string      { return agentVersion }
