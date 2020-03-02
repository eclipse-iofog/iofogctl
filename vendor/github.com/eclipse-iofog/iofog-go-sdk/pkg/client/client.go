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
	"strings"
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
		endpoint = endpoint + ":" + ControllerPortString
	}
	return &Client{
		endpoint: endpoint,
		baseURL:  fmt.Sprintf("http://%s%s", endpoint, apiPrefix),
	}
}

func NewAndLogin(endpoint, email, password string) (clt *Client, err error) {
	clt = New(endpoint)
	if err = clt.Login(LoginRequest{Email: email, Password: password}); err != nil {
		return
	}
	return
}

func NewWithToken(endpoint, token string) (clt *Client, err error) {
	clt = New(endpoint)
	clt.SetAccessToken(token)
	return
}

func (clt *Client) GetEndpoint() string {
	return clt.endpoint
}

func (clt *Client) GetAccessToken() string {
	return clt.accessToken
}

func (clt *Client) SetAccessToken(token string) {
	clt.accessToken = token
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

func (clt *Client) isLoggedIn() bool {
	return clt.accessToken != ""
}
