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

func (clt *Client) CreateApplicationTemplate(request *ApplicationTemplateCreateRequest) (*ApplicationTemplateCreateResponse, error) {
	// Check capability
	if err := clt.IsApplicationTemplateCapable(); err != nil {
		return nil, err
	}

	// Run request
	response := ApplicationTemplateCreateResponse{}
	body, err := clt.doRequest("POST", "/applicationTemplate", request)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (clt *Client) UpdateApplicationTemplate(request *ApplicationTemplateUpdateRequest) (*ApplicationTemplateUpdateResponse, error) {
	// Check capability
	if err := clt.IsApplicationTemplateCapable(); err != nil {
		return nil, err
	}

	// Run request
	response := ApplicationTemplateUpdateResponse{}
	url := fmt.Sprintf("/applicationTemplate/%s", request.Name)
	body, err := clt.doRequest("PUT", url, request)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (clt *Client) UpdateApplicationTemplateMetadata(name string, newMeta *ApplicationTemplateMetadataUpdateRequest) error {
	// Check capability
	if err := clt.IsApplicationTemplateCapable(); err != nil {
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
	// Check capability
	if err := clt.IsApplicationTemplateCapable(); err != nil {
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
	// Check capability
	if err := clt.IsApplicationTemplateCapable(); err != nil {
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
	// Check capability
	if err := clt.IsApplicationTemplateCapable(); err != nil {
		return err
	}

	// Run request
	url := fmt.Sprintf("/applicationTemplate/%s", name)
	if _, err := clt.doRequest("DELETE", url, nil); err != nil {
		return err
	}
	return nil
}
