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

// GetFlowByID retrieve flow information using the Controller REST API
func (clt *Client) GetFlowByID(ID int) (flow *FlowInfo, err error) {
	body, err := clt.doRequest("GET", fmt.Sprintf("/flow/%d", ID), nil)
	if err != nil {
		return
	}
	flow = new(FlowInfo)
	if err = json.Unmarshal(body, flow); err != nil {
		return
	}
	return
}

// CreateFlow creates a new flow using the Controller REST API
func (clt *Client) CreateFlow(name, description string) (*FlowInfo, error) {
	response := FlowCreateResponse{}
	body, err := clt.doRequest("POST", "/flow", FlowCreateRequest{Name: name, Description: description})
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	return clt.GetFlowByID(response.ID)
}

// UpdateFlow patches a flow using the Controller REST API
func (clt *Client) UpdateFlow(request *FlowUpdateRequest) (*FlowInfo, error) {
	_, err := clt.doRequest("PATCH", fmt.Sprintf("/flow/%d", request.ID), *request)
	if err != nil {
		return nil, err
	}
	return clt.GetFlowByID(request.ID)
}

// StartFlow set the flow as active using the Controller REST API
func (clt *Client) StartFlow(ID int) (*FlowInfo, error) {
	active := true
	return clt.UpdateFlow(&FlowUpdateRequest{ID: ID, IsActivated: &active})
}

// StopFlow set the flow as inactive using the Controller REST API
func (clt *Client) StopFlow(ID int) (*FlowInfo, error) {
	active := false
	return clt.UpdateFlow(&FlowUpdateRequest{ID: ID, IsActivated: &active})
}

// GetAllFlows retrieve all flows information from the Controller REST API
func (clt *Client) GetAllFlows() (response *FlowListResponse, err error) {
	body, err := clt.doRequest("GET", "/flow", nil)
	if err != nil {
		return
	}
	response = new(FlowListResponse)
	if err = json.Unmarshal(body, response); err != nil {
		return
	}
	return response, nil
}

// GetFlowByName retrieve the flow information by getting all flows then searching for the first occurance in the list
func (clt *Client) GetFlowByName(name string) (_ *FlowInfo, err error) {
	list, err := clt.GetAllFlows()
	if err != nil {
		return
	}
	for _, flow := range list.Flows {
		if flow.Name == name {
			return &flow, nil
		}
	}
	return nil, NewNotFoundError(fmt.Sprintf("Could not find flow: %s", name))
}

// DeleteFlow deletes a flow using the Controller REST API
func (clt *Client) DeleteFlow(ID int) (err error) {
	_, err = clt.doRequest("DELETE", fmt.Sprintf("/flow/%d", ID), nil)
	return
}
