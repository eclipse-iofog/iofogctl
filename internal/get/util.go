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

package get

import (
	"regexp"
	"strings"

	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

func getEndpointAndPort(connectionString, defaultPort string) (endpoint, port string) {
	// Remove prefix
	endpoint = connectionString
	regex := regexp.MustCompile("https?://")
	port = regex.ReplaceAllString(endpoint, "")

	if !strings.Contains(port, ":") {
		// No port, return connectionString as endpoint and default port
		port = defaultPort
		endpoint = connectionString
	} else {
		// Port specified, extract port and return connection string - port as endpoint
		port = util.After(port, ":")
		endpoint = connectionString[:len(connectionString)-(len(port)+1)]
	}
	return
}
