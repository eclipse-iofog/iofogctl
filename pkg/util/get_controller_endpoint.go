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
	"net/url"
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
)

func GetControllerEndpoint(host string) (endpoint string, err error) {
	// Generate controller endpoint
	u, err := url.Parse(host)
	if err != nil || u.Host == "" {

		if !strings.Contains(host, ":") {
			host = host + ":" + client.ControllerPortString
		}

		u, err = url.Parse("//" + host) // Try to see if controllerEndpoint is an IP, in which case it needs to be pefixed by //
		if err != nil {
			return "", err
		}
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	return u.String(), nil
}
