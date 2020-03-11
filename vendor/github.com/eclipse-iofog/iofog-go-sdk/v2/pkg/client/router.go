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

func (clt *Client) PutDefaultRouter(router Router) (err error) {
	// Send request
	_, err = clt.doRequest("PUT", "/router", router)
	return err
}

func (clt *Client) GetDefaultRouter() (router Router, err error) {
	// Send request
	body, err := clt.doRequest("GET", "/router", nil)
	if err != nil {
		return
	}

	if err = json.Unmarshal(body, &router); err != nil {
		return
	}
	return
}
