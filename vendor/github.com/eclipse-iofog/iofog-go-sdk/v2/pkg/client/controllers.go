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

func (clt *Client) GetStatus() (status ControllerStatus, err error) {
	// Prepare request
	body, err := clt.doRequest("GET", "/status", nil)
	if err != nil {
		return
	}

	// Return body
	if err = json.Unmarshal(body, &status); err != nil {
		return
	}
	return
}
