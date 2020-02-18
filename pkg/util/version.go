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
	versionNumber  = "undefined"
	platform       = "undefined"
	commit         = "undefined"
	date           = "undefined"
	controllerTag  = "undefined"
	kubeletTag     = "undefined"
	proxyTag       = "undefined"
	operatorTag    = "undefined"
	portManagerTag = "undefined"
	agentTag       = "undefined"
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
func GetKubeletTag() string     { return kubeletTag }
func GetProxyTag() string       { return proxyTag }
func GetOperatorTag() string    { return operatorTag }
func GetPortManagerTag() string { return portManagerTag }
func GetAgentTag() string       { return agentTag }
