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

import (
	"fmt"

	rice "github.com/GeertJohan/go.rice"
)

var assets *rice.Box

func init() {
	var err error
	assets, err = rice.FindBox("../../assets")
	Check(err)
}

func GetStaticFileOrDie(filename string) string {
	staticFile, err := GetStaticFile(filename)
	Check(err)
	return staticFile
}

func GetStaticFile(filename string) (string, error) {
	fileContent, err := assets.String(filename)
	if err != nil {
		msg := "could not load static file %s: %s"
		err = fmt.Errorf(msg, filename, err.Error())
		return "", err
	}
	return fileContent, nil
}
