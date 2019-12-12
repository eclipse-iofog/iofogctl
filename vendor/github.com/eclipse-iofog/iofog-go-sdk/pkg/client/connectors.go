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
	"strings"
)

func (clt *Client) ListConnectors() (response ConnectorInfoList, err error) {
	if !clt.isLoggedIn() {
		err = NewError("Controller client must be logged into perform Get Connectors request")
		return
	}

	// Send request
	body, err := clt.doRequest("GET", "/connector", nil)
	if err != nil {
		return
	}

	// Return body
	if err = json.Unmarshal(body, &response); err != nil {
		return
	}

	return
}

func (clt *Client) DeleteConnector(name string) (err error) {
	if !clt.isLoggedIn() {
		return NewError("Controller client must be logged into perform Delete Connector request")
	}

	// Send request
	if _, err = clt.doRequest("DELETE", "/connector", ConnectorInfo{Name: name}); err != nil {
		return
	}

	return
}

func (clt *Client) AddConnector(request ConnectorInfo) error {
	if !clt.isLoggedIn() {
		return NewError("Controller client must be logged into perform Add Connector request")
	}

	// Send request
	_, err := clt.doRequest("POST", "/connector", request)
	if err != nil {
		// Retry with a PUT if already exists
		httpErr, ok := err.(*(HTTPError))
		if ok == true && httpErr.Code == 400 && strings.Contains(httpErr.Error(), "Model already exists") {
			_, err = clt.doRequest("PUT", "/connector", request)
		}
	}
	return err
}

func (clt *Client) UpdateConnector(request ConnectorInfo) error {
	if !clt.isLoggedIn() {
		return NewError("Controller client must be logged into perform Add Connector request")
	}
	// Send request
	_, err := clt.doRequest("PUT", "/connector", request)
	if err != nil {
		return err
	}
	return nil
}

func (clt *Client) isLoggedIn() bool {
	return clt.accessToken != ""
}
