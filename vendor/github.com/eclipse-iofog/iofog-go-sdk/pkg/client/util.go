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

package client

import (
	"bytes"
	"fmt"
	"io"
)

// AgentTypeAgentTypeIDDict Map from string agent type to numeric id
var AgentTypeAgentTypeIDDict = map[string]int{
	"x86": 1,
	"arm": 2,
}

// AgentTypeIDAgentTypeDict Map from numeric id agent type to string agent type
var AgentTypeIDAgentTypeDict = map[int]string{
	1: "x86",
	2: "arm",
}

// RegistryTypeRegistryTypeIDDict Map from string registry type to numeric id
var RegistryTypeRegistryTypeIDDict = map[string]int{
	"remote": 1,
	"local":  2,
}

// RegistryTypeIDRegistryTypeDict Map from numeric id registry type to string
var RegistryTypeIDRegistryTypeDict = map[int]string{
	1: "remote",
	2: "local",
}

func getString(in io.Reader) (out string, err error) {
	buf := new(bytes.Buffer)
	if _, err = buf.ReadFrom(in); err != nil {
		return
	}

	out = buf.String()
	return
}

func checkStatusCode(code int, method, url string, body io.Reader) error {
	if code < 200 || code >= 300 {
		bodyString, err := getString(body)
		if err != nil {
			return err
		}
		return NewHTTPError(fmt.Sprintf("Received %d from %s %s\n%s", code, method, url, bodyString), code)
	}
	return nil
}
