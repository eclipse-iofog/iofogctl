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

package cmd

import (
	"fmt"
	"strings"
)

var pkg struct {
	flagDescDetached string
	flagDescYaml     string
	succRename       string
	succMove         string
}

func init() {
	pkg.flagDescDetached = "Specify command is to run against detached resources"
	pkg.flagDescYaml = "YAML file containing specifications for ioFog resources to deploy"
	pkg.succRename = "Successfully renamed %s %s to %s"
	pkg.succMove = "Successfully moved %s %s to %s %s"
}

func getRenameSuccessMessage(resource, oldName, newName string) string {
	return fmt.Sprintf(pkg.succRename, strings.Title(strings.ToLower(resource)), oldName, newName)
}

func getMoveSuccessMessage(resource, name, otherResource, otherName string) string {
	return fmt.Sprintf(pkg.succRename, resource, name, otherResource, otherName)
}
