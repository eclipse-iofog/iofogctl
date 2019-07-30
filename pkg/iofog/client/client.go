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
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
)

type Client struct {
	endpoint    string
	baseURL     string
	accessToken string
}

var apiPrefix = "/api/v3"

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
		baseURL:  fmt.Sprintf("http://%s%s", endpoint, apiPrefix),
	}
}

func (clt *Client) GetEndpoint() string {
	return clt.endpoint
}

func (clt *Client) makeRequestUrl(url string) string {
	if !strings.HasPrefix(url, "/") {
		url = "/" + url
	}
	return clt.baseURL + url
}

func (clt *Client) doRequest(method, url string, request interface{}) ([]byte, error) {
	// Prepare request
	requestURL := clt.makeRequestUrl(url)
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": clt.accessToken,
	}

	// Send request
	return httpDo(method, requestURL, headers, request)
}
