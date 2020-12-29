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

// GetApplicationByName retrieve application information using the Controller REST API
func (clt *Client) GetApplicationByName(name string) (application *ApplicationInfo, err error) {
	body, err := clt.doRequest("GET", fmt.Sprintf("/application/%s", name), nil)
	if err != nil {
		return
	}
	application = new(ApplicationInfo)
	if err = json.Unmarshal(body, application); err != nil {
		return
	}
	return
}

// CreateApplication creates a new application using the Controller REST API
func (clt *Client) CreateApplication(request *ApplicationCreateRequest) (*ApplicationInfo, error) {
	response := FlowCreateResponse{}
	body, err := clt.doRequest("POST", "/application", request)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	return clt.GetApplicationByName(request.Name)
}

// UpdateApplication updates an application using the Controller REST API
func (clt *Client) UpdateApplication(name string, request *ApplicationUpdateRequest) (*ApplicationInfo, error) {
	_, err := clt.doRequest("PUT", fmt.Sprintf("/application/%s", name), *request)
	if err != nil {
		return nil, err
	}
	return clt.GetApplicationByName(name)
}

// UpdateApplication patches an application using the Controller REST API
func (clt *Client) PatchApplication(name string, request *ApplicationPatchRequest) (*ApplicationInfo, error) {
	_, err := clt.doRequest("PATCH", fmt.Sprintf("/application/%s", name), *request)
	if err != nil {
		return nil, err
	}
	return clt.GetApplicationByName(name)
}

// StartApplication set the application as active using the Controller REST API
func (clt *Client) StartApplication(name string) (*ApplicationInfo, error) {
	active := true
	return clt.PatchApplication(name, &ApplicationPatchRequest{IsActivated: &active})
}

// StopApplication set the application as inactive using the Controller REST API
func (clt *Client) StopApplication(name string) (*ApplicationInfo, error) {
	active := false
	return clt.PatchApplication(name, &ApplicationPatchRequest{IsActivated: &active})
}

// GetAllApplications retrieve all flows information from the Controller REST API
func (clt *Client) GetAllApplications() (response *ApplicationListResponse, err error) {
	body, err := clt.doRequest("GET", "/application", nil)
	if err != nil {
		return
	}
	response = new(ApplicationListResponse)
	if err = json.Unmarshal(body, response); err != nil {
		return
	}
	return response, nil
}

// DeleteApplication deletes an application using the Controller REST API
func (clt *Client) DeleteApplication(name string) (err error) {
	_, err = clt.doRequest("DELETE", fmt.Sprintf("/application/%s", name), nil)
	return
}
