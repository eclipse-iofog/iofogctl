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

// GetCatalog retrieves all catalog items using Controller REST API
func (clt *Client) GetCatalog() (response *CatalogListResponse, err error) {
	body, err := clt.doRequest("GET", fmt.Sprintf("/catalog/microservices"), nil)
	if err != nil {
		return
	}

	response = new(CatalogListResponse)
	err = json.Unmarshal(body, response)
	return
}

// GetCatalogItem retrieves one catalog item using Controller REST API
func (clt *Client) GetCatalogItem(ID int) (response *CatalogItemInfo, err error) {
	body, err := clt.doRequest("GET", fmt.Sprintf("/catalog/microservices/%d", ID), nil)
	if err != nil {
		return
	}

	response = new(CatalogItemInfo)
	err = json.Unmarshal(body, response)
	return
}

// CreateCatalogItem creates one catalog item using Controller REST API
func (clt *Client) CreateCatalogItem(request *CatalogItemCreateRequest) (*CatalogItemInfo, error) {
	// Set registry to public docker by default
	if request.RegistryID == 0 {
		request.RegistryID = 1
	}

	body, err := clt.doRequest("POST", fmt.Sprintf("/catalog/microservices"), request)
	if err != nil {
		return nil, err
	}
	response := &CatalogItemCreateResponse{}
	if err = json.Unmarshal(body, response); err != nil {
		return nil, err
	}
	return clt.GetCatalogItem(response.ID)
}

// DeleteCatalogItem deletes one catalog item using Controller REST API
func (clt *Client) DeleteCatalogItem(ID int) (err error) {
	_, err = clt.doRequest("DELETE", fmt.Sprintf("/catalog/microservices/%d", ID), nil)
	return
}