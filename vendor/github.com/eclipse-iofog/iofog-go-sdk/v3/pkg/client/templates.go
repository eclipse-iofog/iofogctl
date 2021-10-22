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

func (clt *Client) IsApplicationTemplateCapable() error {
	if _, err := clt.doRequest("HEAD", "/capabilities/applicationTemplates", nil); err != nil {
		// If 404, not capable
		if _, ok := err.(*NotFoundError); ok {
			return NewNotSupportedError("Application Templates")
		}
		return err
	}
	return nil
}

func (clt *Client) applicationTemplatePreflight() error {
	// Check capability
	if err := clt.IsApplicationTemplateCapable(); err != nil {
		return err
	}

	if !clt.isLoggedIn() {
		return NewError(edgeResourceLoggedInErr)
	}

	return nil
}

// CreateApplicationTemplateFromYAML creates a new application template using the Controller REST API
// It sends the yaml file to Controller REST API
func (clt *Client) CreateApplicationTemplateFromYAML(file io.Reader) (*ApplicationTemplate, error) {
	requestBody := &bytes.Buffer{}
	writer := multipart.NewWriter(requestBody)
	part, _ := writer.CreateFormFile("template", "application.yaml")
	_, err := io.Copy(part, file)
	if err != nil {
		return nil, err
	}
	writer.Close()

	headers := map[string]string{
		"Content-Type": writer.FormDataContentType(),
	}
	body, err := clt.doRequestWithHeaders("POST", "/applicationTemplate/yaml", requestBody, headers)

	if err != nil {
		return nil, err
	}
	response := ApplicationTemplateCreateResponse{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	return clt.GetApplicationTemplate(response.Name)
}

// UpdateApplicationTemplateFromYAML updates an application template using the Controller REST API
// It sends the yaml file to Controller REST API
func (clt *Client) UpdateApplicationTemplateFromYAML(name string, file io.Reader) (*ApplicationTemplate, error) {
	requestBody := &bytes.Buffer{}
	writer := multipart.NewWriter(requestBody)
	part, _ := writer.CreateFormFile("template", "microservice.yaml")
	_, err := io.Copy(part, file)
	if err != nil {
		return nil, err
	}
	writer.Close()

	headers := map[string]string{
		"Content-Type": writer.FormDataContentType(),
	}

	_, err = clt.doRequestWithHeaders("PUT", fmt.Sprintf("/applicationTemplate/yaml/%s", name), requestBody, headers)
	if err != nil {
		return nil, err
	}
	return clt.GetApplicationTemplate(name)
}

func (clt *Client) UpdateApplicationTemplateMetadata(name string, newMeta *ApplicationTemplateMetadataUpdateRequest) error {
	if err := clt.applicationTemplatePreflight(); err != nil {
		return err
	}

	// Run request
	url := fmt.Sprintf("/applicationTemplate/%s", name)
	if _, err := clt.doRequest("PATCH", url, newMeta); err != nil {
		return err
	}
	return nil
}

func (clt *Client) ListApplicationTemplates() (*ApplicationTemplateListResponse, error) {
	if err := clt.applicationTemplatePreflight(); err != nil {
		return nil, err
	}

	// Run request
	response := ApplicationTemplateListResponse{}
	body, err := clt.doRequest("GET", "/applicationTemplates", nil)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (clt *Client) GetApplicationTemplate(name string) (*ApplicationTemplate, error) {
	if err := clt.applicationTemplatePreflight(); err != nil {
		return nil, err
	}

	// Run request
	response := ApplicationTemplate{}
	url := fmt.Sprintf("/applicationTemplate/%s", name)
	body, err := clt.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (clt *Client) DeleteApplicationTemplate(name string) error {
	if err := clt.applicationTemplatePreflight(); err != nil {
		return err
	}

	// Run request
	url := fmt.Sprintf("/applicationTemplate/%s", name)
	if _, err := clt.doRequest("DELETE", url, nil); err != nil {
		return err
	}
	return nil
}
