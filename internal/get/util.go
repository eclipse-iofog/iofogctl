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

	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

func getAddressAndPort(endpoint, defaultPort string) (addr, port string) {
	port = util.AfterLast(endpoint, ":")
	if port == "" {
		port = defaultPort
	}
	// Remove prefix
	regex := regexp.MustCompile("https?://")
	addr = regex.ReplaceAllString(endpoint, "")
	addr = util.Before(addr, ":")

	return
}
