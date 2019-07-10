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
	branch        = "undefined"
	commit        = "undefined"
	date          = "undefined"
	platform      = "undefined"
)

const (
	LocalBuildVersion = "dev"
	DevVersionSuffix  = "-b"
)

type Version struct {
	VersionNumber string `yaml:"version"`
	Branch        string
	Commit        string
	Date          string
	Platform      string
}

func GetVersion() Version {
	return Version{
		VersionNumber: versionNumber,
		Branch:        branch,
		Commit:        commit,
		Date:          date,
		Platform:      platform,
	}
}
