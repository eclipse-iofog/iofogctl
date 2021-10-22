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
	"io"
	"mime/multipart"
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

// CreateApplicationFromYAML creates a new application using the Controller REST API
// It sends the yaml file to Controller REST API
func (clt *Client) CreateApplicationFromYAML(file io.Reader) (*ApplicationInfo, error) {
	requestBody := &bytes.Buffer{}
	writer := multipart.NewWriter(requestBody)
	part, _ := writer.CreateFormFile("application", "application.yaml")
	_, err := io.Copy(part, file)
	if err != nil {
		return nil, err
	}
	writer.Close()

	headers := map[string]string{
		"Content-Type": writer.FormDataContentType(),
	}
	body, err := clt.doRequestWithHeaders("POST", "/application/yaml", requestBody, headers)

	if err != nil {
		return nil, err
	}
	response := FlowCreateResponse{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	return clt.GetApplicationByName(response.Name)
}

// UpdateApplicationFromYAML updates an application using the Controller REST API
// It sends the yaml file to Controller REST API
func (clt *Client) UpdateApplicationFromYAML(name string, file io.Reader) (*ApplicationInfo, error) {
	requestBody := &bytes.Buffer{}
	writer := multipart.NewWriter(requestBody)
	part, _ := writer.CreateFormFile("application", "application.yaml")
	_, err := io.Copy(part, file)
	if err != nil {
		return nil, err
	}
	writer.Close()

	headers := map[string]string{
		"Content-Type": writer.FormDataContentType(),
	}

	_, err = clt.doRequestWithHeaders("PUT", fmt.Sprintf("/application/yaml/%s", name), requestBody, headers)
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
	newName := name
	if request.Name != nil {
		newName = *request.Name
	}
	return clt.GetApplicationByName(newName)
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
