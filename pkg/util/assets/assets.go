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

package assets

import (
	"fmt"

	rice "github.com/GeertJohan/go.rice"

	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

var assets *rice.Box

func init() {
	var err error
	assets, err = rice.FindBox("../../../assets")
	util.Check(err)
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
