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
	controllerTag = "undefined"
	connectorTag  = "undefined"
	kubeletTag    = "undefined"
	operatorTag   = "undefined"
	schedulerTag  = "undefined"
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

func GetControllerTag() string {
	return controllerTag
}

func GetConnectorTag() string {
	return connectorTag
}

func GetKubeletTag() string {
	return kubeletTag
}

func GetOperatorTag() string {
	return operatorTag
}

func GetSchedulerTag() string {
	return schedulerTag
}
