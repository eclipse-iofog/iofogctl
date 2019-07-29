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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Client struct {
	endpoint    string
	baseURL     string
	accessToken string
}

func New(endpoint string) *Client {
	// Remove prefix
	regex := regexp.MustCompile("https?://")
	endpoint = regex.ReplaceAllString(endpoint, "")

	// Add default port if none specified
	if !strings.Contains(endpoint, ":") {
		endpoint = endpoint + ":" + strconv.Itoa(iofog.ControllerPort)
	}
	return &Client{
		endpoint: endpoint,
		baseURL:  fmt.Sprintf("http://%s/api/v3/", endpoint),
	}
}

func (this *Client) GetEndpoint() string {
	return this.endpoint
}

func (this *Client) GetStatus() (status ControllerStatus, err error) {
	// Prepare request
	method := "GET"
	url := this.baseURL + "status"

	// Send request
	body, err := httpDo(method, url, nil, nil)
	if err != nil {
		return
	}

	// Return body
	if err = json.Unmarshal(body, &status); err != nil {
		return
	}
	return
}

func (this *Client) CreateUser(request User) error {
	// Prepare request
	method := "POST"
	url := this.baseURL + "user/signup"
	headers := map[string]string{"Content-Type": "application/json"}

	// Send request
	if _, err := httpDo(method, url, headers, request); err != nil {
		return err
	}

	return nil
}

func (this *Client) Login(request LoginRequest) (err error) {
	// Prepare request
	method := "POST"
	url := this.baseURL + "user/login"
	headers := map[string]string{"Content-Type": "application/json"}

	// Send request
	body, err := httpDo(method, url, headers, request)
	if err != nil {
		return
	}

	// Read access token from response
	var response LoginResponse
	if err = json.Unmarshal(body, &response); err != nil {
		return
	}
	this.accessToken = response.AccessToken

	return
}

func (this *Client) CreateAgent(request CreateAgentRequest) (response CreateAgentResponse, err error) {
	if !this.isLoggedIn() {
		err = util.NewError("Controller client must be logged into perform Create Agent request")
		return
	}

	// Prepare request
	method := "POST"
	url := this.baseURL + "iofog"
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": this.accessToken,
	}

	// Send request
	body, err := httpDo(method, url, headers, request)
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

func (this *Client) GetAgentProvisionKey(UUID string) (response GetAgentProvisionKeyResponse, err error) {
	if !this.isLoggedIn() {
		err = util.NewError("Controller client must be logged into perform Get Agent Provisioning Key request")
		return
	}

	// Prepare request
	method := "GET"
	url := this.baseURL + "iofog/" + UUID + "/provisioning-key"
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": this.accessToken,
	}

	// Send request
	body, err := httpDo(method, url, headers, nil)
	if err != nil {
		return
	}

	if err = json.Unmarshal(body, &response); err != nil {
		return
	}
	return
}

func (this *Client) ListAgents() (response ListAgentsResponse, err error) {
	if !this.isLoggedIn() {
		err = util.NewError("Controller client must be logged into perform List Agents request")
		return
	}

	// Prepare request
	method := "GET"
	url := this.baseURL + "iofog-list"
	headers := map[string]string{
		"Authorization": this.accessToken,
	}

	// Send request
	body, err := httpDo(method, url, headers, AgentListFilter{})
	if err != nil {
		return
	}

	// Return body
	if err = json.Unmarshal(body, &response); err != nil {
		return
	}

	return
}

func (this *Client) GetAgent(UUID string) (response AgentInfo, err error) {
	if !this.isLoggedIn() {
		err = util.NewError("Controller client must be logged into perform Get Agent request")
		return
	}

	// Prepare request
	method := "GET"
	url := this.baseURL + "iofog/" + UUID
	headers := map[string]string{
		"Authorization": this.accessToken,
	}

	// Send request
	body, err := httpDo(method, url, headers, nil)
	if err != nil {
		return
	}

	// Return body
	if err = json.Unmarshal(body, &response); err != nil {
		return
	}

	return
}

func (this *Client) DeleteAgent(UUID string) error {
	if !this.isLoggedIn() {
		return util.NewError("Controller client must be logged into perform Delete Agent request")
	}

	// Prepare request
	method := "DELETE"
	url := this.baseURL + "iofog/" + UUID
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": this.accessToken,
	}

	// Send request
	if _, err := httpDo(method, url, headers, nil); err != nil {
		return err
	}

	return nil
}

func (this *Client) GetConnectors() (response ConnectorInfoList, err error) {
	if !this.isLoggedIn() {
		err = util.NewError("Controller client must be logged into perform Get Connectors request")
		return
	}

	// Prepare request
	method := "GET"
	url := this.baseURL + "connector"

	// Send request
	body, err := httpDo(method, url, nil, nil)
	if err != nil {
		return
	}

	// Return body
	if err = json.Unmarshal(body, &response); err != nil {
		return
	}

	return
}

func (this *Client) DeleteConnector(ip string) (err error) {
	if !this.isLoggedIn() {
		return util.NewError("Controller client must be logged into perform Delete Connector request")
	}

	// Prepare request
	method := "DELETE"
	url := this.baseURL + "connector"
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": this.accessToken,
	}
	connectorInfo := ConnectorInfo{IP: ip}

	// Send request
	if _, err = httpDo(method, url, headers, connectorInfo); err != nil {
		return
	}

	return
}

func (this *Client) AddConnector(request ConnectorInfo) error {
	if !this.isLoggedIn() {
		return util.NewError("Controller client must be logged into perform Add Connector request")
	}

	// Prepare request
	contentType := "application/json"
	url := this.baseURL + "connector"
	body, err := json.Marshal(request)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequest("POST", url, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Authorization", this.accessToken)
	httpReq.Header.Set("Content-Type", contentType)

	// Send request
	client := http.Client{}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return err
	}

	// Read the body
	buf := new(bytes.Buffer)
	buf.ReadFrom(httpResp.Body)

	// Retry with a PUT if already exists
	if httpResp.StatusCode == 400 && strings.Contains(buf.String(), "Model already exists") {
		httpReq.Method = "PUT"
		httpResp, err = client.Do(httpReq)
		if err != nil {
			return err
		}
	}

	// Check response code
	if err = checkStatusCode(httpResp.StatusCode, httpReq.Method, httpReq.URL.String(), httpResp.Body); err != nil {
		return err
	}

	return nil
}

func (this *Client) isLoggedIn() bool {
	return this.accessToken != ""
}
