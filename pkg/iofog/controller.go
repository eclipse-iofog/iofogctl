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

package iofog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"io"
	"net/http"
	"regexp"
	"strings"
)

type Controller struct {
	baseURL string
}

func NewController(endpoint string) *Controller {
	regex := regexp.MustCompile("https?://")
	endpoint = regex.ReplaceAllString(endpoint, "")
	return &Controller{
		baseURL: fmt.Sprintf("http://%s/api/v3/", endpoint),
	}
}

func (ctrl *Controller) GetStatus() (status ControllerStatus, err error) {
	url := ctrl.baseURL + "status"
	httpResp, err := http.Get(url)
	if err != nil {
		return
	}

	if err = checkStatusCode(httpResp.StatusCode, "GET", url, httpResp.Body); err != nil {
		return
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(httpResp.Body)
	err = json.Unmarshal(buf.Bytes(), &status)
	if err != nil {
		return
	}
	return
}

func (ctrl *Controller) CreateUser(request User) error {
	// Prepare request
	contentType := "application/json"
	url := ctrl.baseURL + "user/signup"
	body, err := json.Marshal(request)
	if err != nil {
		return err
	}

	// Send request
	httpResp, err := http.Post(url, contentType, strings.NewReader(string(body)))
	if err != nil {
		return err
	}

	// Check response code
	return checkStatusCode(httpResp.StatusCode, "POST", url, httpResp.Body)
}

func (ctrl *Controller) Login(request LoginRequest) (response LoginResponse, err error) {
	// Prepare request
	contentType := "application/json"
	url := ctrl.baseURL + "user/login"
	body, err := json.Marshal(request)
	if err != nil {
		return
	}

	// Send request
	httpResp, err := http.Post(url, contentType, strings.NewReader(string(body)))
	if err != nil {
		return
	}

	// Check response code
	if err = checkStatusCode(httpResp.StatusCode, "POST", url, httpResp.Body); err != nil {
		return
	}

	// Read response body
	buf := new(bytes.Buffer)
	buf.ReadFrom(httpResp.Body)
	err = json.Unmarshal(buf.Bytes(), &response)
	if err != nil {
		return
	}

	return
}

func (ctrl *Controller) CreateAgent(request CreateAgentRequest, accessToken string) (response CreateAgentResponse, err error) {
	// Prepare request
	method := "POST"
	contentType := "application/json"
	url := ctrl.baseURL + "iofog"
	body, err := json.Marshal(request)
	if err != nil {
		return
	}
	httpReq, err := http.NewRequest(method, url, strings.NewReader(string(body)))
	if err != nil {
		return
	}
	httpReq.Header.Set("Authorization", accessToken)
	httpReq.Header.Set("Content-Type", contentType)

	// Send request
	client := http.Client{}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return
	}

	// Check response code
	if err = checkStatusCode(httpResp.StatusCode, method, url, httpResp.Body); err != nil {
		return
	}

	// TODO: Determine full type returned from this endpoint
	// Read uuid from response
	var respMap map[string]interface{}
	buf := new(bytes.Buffer)
	buf.ReadFrom(httpResp.Body)
	err = json.Unmarshal(buf.Bytes(), &respMap)
	if err != nil {
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

func (ctrl *Controller) GetAgentProvisionKey(UUID, accessToken string) (response GetAgentProvisionKeyResponse, err error) {
	// Prepare request
	method := "GET"
	contentType := "application/json"
	url := ctrl.baseURL + "iofog/" + UUID + "/provisioning-key"
	body := strings.NewReader("")
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", accessToken)
	req.Header.Set("Content-Type", contentType)

	// Send request
	client := http.Client{}
	httpResp, err := client.Do(req)
	if err != nil {
		return
	}

	// Check response code
	if err = checkStatusCode(httpResp.StatusCode, method, url, httpResp.Body); err != nil {
		return
	}

	// Read body
	buf := new(bytes.Buffer)
	buf.ReadFrom(httpResp.Body)
	err = json.Unmarshal(buf.Bytes(), &response)
	if err != nil {
		return
	}
	return
}

func (ctrl *Controller) ListAgents(accessToken string) (response ListAgentsResponse, errr error) {
	// Prepare request
	method := "GET"
	url := ctrl.baseURL + "iofog-list"
	filter := AgentListFilter{}
	body, err := json.Marshal(filter)
	if err != nil {
		return
	}
	httpReq, err := http.NewRequest(method, url, strings.NewReader(string(body)))
	if err != nil {
		return
	}
	httpReq.Header.Set("Authorization", accessToken)

	// Send request
	client := http.Client{}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return
	}

	// Check response code
	if err = checkStatusCode(httpResp.StatusCode, method, url, httpResp.Body); err != nil {
		return
	}

	// Read body
	buf := new(bytes.Buffer)
	buf.ReadFrom(httpResp.Body)
	err = json.Unmarshal(buf.Bytes(), &response)
	if err != nil {
		return
	}

	return
}

func (ctrl *Controller) GetAgent(UUID, accessToken string) (response AgentInfo, err error) {
	// Prepare request
	method := "GET"
	url := ctrl.baseURL + "iofog/" + UUID
	body := strings.NewReader("")
	httpReq, err := http.NewRequest(method, url, body)
	if err != nil {
		return
	}
	httpReq.Header.Set("Authorization", accessToken)

	// Send request
	client := http.Client{}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return
	}

	// Check response code
	if err = checkStatusCode(httpResp.StatusCode, method, url, httpResp.Body); err != nil {
		return
	}

	// Read body
	buf := new(bytes.Buffer)
	buf.ReadFrom(httpResp.Body)
	err = json.Unmarshal(buf.Bytes(), &response)
	if err != nil {
		return
	}

	return
}

func (ctrl *Controller) DeleteAgent(UUID, accessToken string) error {
	// Prepare request
	method := "DELETE"
	contentType := "application/json"
	url := ctrl.baseURL + "iofog/" + UUID
	body := strings.NewReader("")
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", accessToken)
	req.Header.Set("Content-Type", contentType)

	// Send request
	client := http.Client{}
	httpResp, err := client.Do(req)
	if err != nil {
		return err
	}

	// Check response code
	if err = checkStatusCode(httpResp.StatusCode, method, url, httpResp.Body); err != nil {
		return err
	}

	return nil
}

func (ctrl *Controller) GetConnectors(accessToken string) (response ConnectorInfoList, err error) {
	// Prepare request
	method := "GET"
	url := ctrl.baseURL + "connector"
	httpReq, err := http.NewRequest(method, url, strings.NewReader(""))
	if err != nil {
		return
	}

	// Send request
	client := http.Client{}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return
	}

	// Check response code
	if err = checkStatusCode(httpResp.StatusCode, method, url, httpResp.Body); err != nil {
		return
	}

	// Read body
	buf := new(bytes.Buffer)
	buf.ReadFrom(httpResp.Body)
	err = json.Unmarshal(buf.Bytes(), &response)
	if err != nil {
		return
	}

	return
}

func (ctrl *Controller) DeleteConnector(ip, accessToken string) (err error) {
	// Prepare request
	method := "DELETE"
	url := ctrl.baseURL + "connector"
	body := fmt.Sprintf(`{"publicIp":"%s"}`, ip)
	httpReq, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		return
	}

	// Send request
	client := http.Client{}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return
	}

	// Check response code
	return checkStatusCode(httpResp.StatusCode, method, url, httpResp.Body)
}

func (ctrl *Controller) AddConnector(request ConnectorInfo, accessToken string) error {
	// Prepare request
	contentType := "application/json"
	url := ctrl.baseURL + "connector"
	body, err := json.Marshal(request)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequest("POST", url, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Authorization", accessToken)
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
		println("FIRST ERR")
		httpReq.Method = "PUT"
		httpResp, err = client.Do(httpReq)
		if err != nil {
			return err
		}
	}

	// Check response code
	if err = checkStatusCode(httpResp.StatusCode, "PUT", url, httpResp.Body); err != nil {
		return err
	}

	return nil
}

func checkStatusCode(code int, method, url string, body io.Reader) error {
	if code < 200 || code >= 300 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(body)
		return util.NewInternalError(fmt.Sprintf("Received %d from %s %s\n%s", code, method, url, buf.String()))
	}
	return nil
}
