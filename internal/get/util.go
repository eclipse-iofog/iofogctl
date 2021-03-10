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

package get

import (
	"regexp"

	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

func getAddressAndPort(endpoint, defaultPort string) (addr, port string) {
	// Remove prefix
	regex := regexp.MustCompile("https?://")
	addr = regex.ReplaceAllString(endpoint, "")

	// Get port from address
	port = util.AfterLast(addr, ":")
	if port == "" {
		port = defaultPort
	}

	// Remove port from address
	addr = util.Before(addr, ":")

	return
}
