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

	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func (clt *Client) CreateAgent(request CreateAgentRequest) (response CreateAgentResponse, err error) {
	if !clt.isLoggedIn() {
		err = util.NewError("Controller client must be logged into perform Create Agent request")
		return
	}

	// Send request
	body, err := clt.doRequest("POST", "/iofog", request)
	if err != nil {
		return
	}

	// TODO: Determine full type returned from this endpoint
	// Read uuid from response
	var respMap map[string]interface{}
	if err = json.Unmarshal(body, &respMap); err != nil {
		return
	}
	uuid, exists := respMap["uuid"].(string)
	if !exists {
		err = util.NewInternalError("Failed to get new Agent UUID from Controller")
		return
	}

	response.UUID = uuid
	return
}

func (clt *Client) GetAgentProvisionKey(UUID string) (response GetAgentProvisionKeyResponse, err error) {
	if !clt.isLoggedIn() {
		err = util.NewError("Controller client must be logged into perform Get Agent Provisioning Key request")
		return
	}

	// Send request
	body, err := clt.doRequest("GET", fmt.Sprintf("/iofog/%s/provisioning-key", UUID), nil)
	if err != nil {
		return
	}

	if err = json.Unmarshal(body, &response); err != nil {
		return
	}
	return
}

func (clt *Client) ListAgents() (response ListAgentsResponse, err error) {
	if !clt.isLoggedIn() {
		err = util.NewError("Controller client must be logged into perform List Agents request")
		return
	}

	// Send request
	body, err := clt.doRequest("GET", "/iofog-list", AgentListFilter{})
	if err != nil {
		return
	}

	// Return body
	if err = json.Unmarshal(body, &response); err != nil {
		return
	}

	return
}

func (clt *Client) GetAgent(UUID string) (response AgentInfo, err error) {
	if !clt.isLoggedIn() {
		err = util.NewError("Controller client must be logged into perform Get Agent request")
		return
	}

	// Send request
	body, err := clt.doRequest("GET", fmt.Sprintf("/iofog/%s", UUID), nil)
	if err != nil {
		return
	}

	// Return body
	if err = json.Unmarshal(body, &response); err != nil {
		return
	}

	return
}

func (clt *Client) DeleteAgent(UUID string) error {
	if !clt.isLoggedIn() {
		return util.NewError("Controller client must be logged into perform Delete Agent request")
	}

	// Send request
	if _, err := clt.doRequest("DELETE", fmt.Sprintf("/iofog/%s", UUID), nil); err != nil {
		return err
	}

	return nil
}
