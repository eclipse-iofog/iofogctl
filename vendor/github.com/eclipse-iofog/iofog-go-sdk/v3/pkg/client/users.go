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

import "encoding/json"

func (clt *Client) CreateUser(request User) error {
	// Send request
	if _, err := clt.doRequest("POST", "/user/signup", request); err != nil {
		return err
	}

	return nil
}

func (clt *Client) Login(request LoginRequest) (err error) {
	// Send request
	body, err := clt.doRequest("POST", "/user/login", request)
	if err != nil {
		return
	}

	// Read access token from response
	var response LoginResponse
	if err = json.Unmarshal(body, &response); err != nil {
		return
	}
	clt.accessToken = response.AccessToken

	return
}

func (clt *Client) UpdateUserPassword(request UpdateUserPasswordRequest) (err error) {
	// Send request
	_, err = clt.doRequest("PATCH", "/user/password", request)
	if err != nil {
		return
	}

	return
}
