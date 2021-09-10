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
	"sync"

	rice "github.com/GeertJohan/go.rice"
)

var assets *rice.Box

var once sync.Once

func GetStaticFile(filename string) (string, error) {
	var err error
	once.Do(func() {
		assets, err = rice.FindBox("../../assets")
	})
	if err != nil {
		msg := "could not initialize assets: %s"
		err = fmt.Errorf(msg, err.Error())
		return "", err
	}
	fileContent, err := assets.String(filename)
	if err != nil {
		msg := "could not load static file %s: %s"
		err = fmt.Errorf(msg, filename, err.Error())
		return "", err
	}
	return fileContent, nil
}
