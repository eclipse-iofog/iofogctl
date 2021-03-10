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

// CreateRegistry creates a new registry using the Controller REST API
func (clt *Client) CreateRegistry(request *RegistryCreateRequest) (int, error) {
	response := RegistryCreateResponse{}
	body, err := clt.doRequest("POST", "/registries", request)
	if err != nil {
		return -1, err
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return -1, err
	}
	return response.ID, nil
}

// UpdateRegistry patches a registry using the Controller REST API
func (clt *Client) UpdateRegistry(request RegistryUpdateRequest) error {
	_, err := clt.doRequest("PATCH", fmt.Sprintf("/registries/%d", request.ID), request)
	if err != nil {
		return err
	}
	return nil
}

// ListRegistries retrieve all registries information from the Controller REST API
func (clt *Client) ListRegistries() (response RegistryListResponse, err error) {
	body, err := clt.doRequest("GET", "/registries", nil)
	if err != nil {
		return
	}
	if err = json.Unmarshal(body, &response); err != nil {
		return
	}
	return response, nil
}

// DeleteRegistry deletes a registry using the Controller REST API
func (clt *Client) DeleteRegistry(id int) (err error) {
	_, err = clt.doRequest("DELETE", fmt.Sprintf("/registries/%d", id), nil)
	return
}
