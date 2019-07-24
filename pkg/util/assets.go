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

import (
	"sync"

	rice "github.com/GeertJohan/go.rice"
)

var once sync.Once

var staticFiles map[string]string

func GetStaticFile(filename string) string {

	once.Do(func() {
		staticFiles = make(map[string]string)
	})

	fileContent, ok := staticFiles[filename]
	if !ok {
		assets, err := rice.FindBox("../../assets")
		Check(err)
		fileContent, err = assets.String(filename)
		Check(err)
		staticFiles[filename] = fileContent
	}
	return fileContent
}
