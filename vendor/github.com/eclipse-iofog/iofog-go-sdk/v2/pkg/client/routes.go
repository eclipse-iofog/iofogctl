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

	body, err := clt.doRequest("GET", fmt.Sprintf("/routes"), nil)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &response)
	return
}

func (clt *Client) GetRoute(name string) (response RouteInfo, err error) {
	if !clt.isLoggedIn() {
		err = NewError("Controller client must be logged into perform Get Route request")
		return
	}

	// Send request
	body, err := clt.doRequest("GET", fmt.Sprintf("/routes/%s", name), nil)
	if err != nil {
		return
	}

	// Return body
	if err = json.Unmarshal(body, &response); err != nil {
		return
	}

	return
}

func (clt *Client) CreateRoute(name, srcMsvcUUID, destMsvcUUID string) (response RouteInfo, err error) {
	if !clt.isLoggedIn() {
		err = NewError("Controller client must be logged into perform Create Route request")
		return
	}

	// Send request
	body, err := clt.doRequest("POST", "/routes", &RouteInfo{
		Name:                   name,
		SourceMicroserviceUUID: srcMsvcUUID,
		DestMicroserviceUUID:   destMsvcUUID,
	})
	if err != nil {
		return
	}

	// Return body
	if err = json.Unmarshal(body, &response); err != nil {
		return
	}

	return
}

func (clt *Client) UpdateRoute(name string, route RouteInfo) (err error) {
	if !clt.isLoggedIn() {
		err = NewError("Controller client must be logged into perform Update Route request")
		return
	}

	// Send request
	if _, err = clt.doRequest("PATCH", fmt.Sprintf("/routes/%s", name), &route); err != nil {
		return
	}

	return
}

func (clt *Client) DeleteRoute(name string) (err error) {
	if !clt.isLoggedIn() {
		err = NewError("Controller client must be logged into perform Delete Route request")
		return
	}

	// Send request
	if _, err = clt.doRequest("DELETE", fmt.Sprintf("/routes/%s", name), nil); err != nil {
		return
	}

	return
}
