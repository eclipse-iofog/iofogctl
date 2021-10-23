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
	"encoding/json"
	"fmt"
)

func (clt *Client) ListRoutes() (response RouteListResponse, err error) {
	if !clt.isLoggedIn() {
		err = NewError("Controller client must be logged into perform List Routes request")
		return
	}

	body, err := clt.doRequest("GET", "/routes", nil)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &response)
	return
}

func (clt *Client) GetRoute(appName, name string) (route Route, err error) {
	if !clt.isLoggedIn() {
		err = NewError("Controller client must be logged into perform Get Route request")
		return
	}

	// Send request
	body, err := clt.doRequest("GET", fmt.Sprintf("/routes/%s/%s", appName, name), nil)
	if err != nil {
		return
	}

	// Return body
	if err = json.Unmarshal(body, &route); err != nil {
		return
	}

	return
}

func (clt *Client) CreateRoute(route *Route) (err error) {
	if !clt.isLoggedIn() {
		err = NewError("Controller client must be logged into perform Create Route request")
		return
	}

	// Send request
	if _, err = clt.doRequest("POST", "/routes", route); err != nil {
		return
	}

	return
}

func (clt *Client) UpdateRoute(route *Route) (err error) {
	if !clt.isLoggedIn() {
		err = NewError("Controller client must be logged into perform Update Route request")
		return
	}

	if _, err = clt.GetRoute(route.Application, route.Name); err == nil {
		return clt.PatchRoute(route.Application, route.Name, route)
	}

	return clt.CreateRoute(route)
}

func (clt *Client) PatchRoute(appName, name string, route *Route) (err error) {
	if !clt.isLoggedIn() {
		err = NewError("Controller client must be logged into perform Update Route request")
		return
	}

	// Send request
	if _, err = clt.doRequest("PATCH", fmt.Sprintf("/routes/%s/%s", appName, name), &route); err != nil {
		return
	}

	return
}

func (clt *Client) DeleteRoute(appName, name string) (err error) {
	if !clt.isLoggedIn() {
		err = NewError("Controller client must be logged into perform Delete Route request")
		return
	}

	// Send request
	if _, err = clt.doRequest("DELETE", fmt.Sprintf("/routes/%s/%s", appName, name), nil); err != nil {
		return
	}

	return
}
