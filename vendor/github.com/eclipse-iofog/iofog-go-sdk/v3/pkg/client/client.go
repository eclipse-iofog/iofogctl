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
	"net/url"
	"path"
	"strings"
	"time"
)

type controllerStatus struct {
	version         string
	versionNoSuffix string
	versionNums     []string
}

type Client struct {
	baseURL     url.URL
	accessToken string
	retries     Retries
	status      controllerStatus
	timeout     int
}

type Options struct {
	BaseURL url.URL
	Retries *Retries
	Timeout int
}

func New(opt Options) *Client {
	if opt.Timeout == 0 {
		opt.Timeout = 5
	}
	retries := GlobalRetriesPolicy
	if opt.Retries != nil {
		retries = *opt.Retries
	}
	client := &Client{
		retries: retries,
		baseURL: opt.BaseURL,
		timeout: opt.Timeout,
	}
	// Get Controller version
	if status, err := client.GetStatus(); err == nil {
		versionNoSuffix := before(status.Versions.Controller, "-")
		versionNums := strings.Split(versionNoSuffix, ".")
		client.status = controllerStatus{
			version:         status.Versions.Controller,
			versionNoSuffix: versionNoSuffix,
			versionNums:     versionNums,
		}
	}
	return client
}

func NewAndLogin(opt Options, email, password string) (clt *Client, err error) {
	clt = New(opt)
	if err = clt.Login(LoginRequest{Email: email, Password: password}); err != nil {
		return
	}
	return
}

func NewWithToken(opt Options, token string) (clt *Client, err error) {
	clt = New(opt)
	clt.SetAccessToken(token)
	return
}

func (clt *Client) GetBaseURL() string {
	return clt.baseURL.String()
}

func (clt *Client) GetRetries() Retries {
	return clt.retries
}

func (clt *Client) SetRetries(retries Retries) {
	clt.retries = retries
}

func (clt *Client) GetAccessToken() string {
	return clt.accessToken
}

func (clt *Client) SetAccessToken(token string) {
	clt.accessToken = token
}

func (clt *Client) doRequestWithRetries(currentRetries Retries, method, requestURL string, headers map[string]string, request interface{}) ([]byte, error) {
	// Send request
	httpDo := httpDo{timeout: clt.timeout}
	bytes, err := httpDo.do(method, requestURL, headers, request)
	if err != nil {
		httpErr, ok := err.(*HTTPError)
		// If HTTP Error
		if ok {
			if httpErr.Code == 408 { // HTTP Timeout
				if currentRetries.Timeout < clt.retries.Timeout {
					currentRetries.Timeout++
					time.Sleep(time.Duration(currentRetries.Timeout) * time.Second)
					return clt.doRequestWithRetries(currentRetries, method, requestURL, headers, request)
				}
				return bytes, err
			}
		}
		// If custom retries defined
		if clt.retries.CustomMessage != nil {
			for message, allowedRetries := range clt.retries.CustomMessage {
				if strings.Contains(err.Error(), message) {
					if currentRetries.CustomMessage[message] < allowedRetries {
						currentRetries.CustomMessage[message]++
						time.Sleep(time.Duration(currentRetries.CustomMessage[message]) * time.Second)
						return clt.doRequestWithRetries(currentRetries, method, requestURL, headers, request)
					}
					return bytes, err
				}
			}
		}
	}
	return bytes, err
}

func (clt *Client) doRequest(method, requestPath string, request interface{}) ([]byte, error) {
	// Prepare request
	requestURL, err := url.Parse(clt.baseURL.String())
	if err != nil {
		return nil, err
	}
	requestURL.Path = path.Join(requestURL.Path, requestPath)
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": clt.accessToken,
	}

	currentRetries := Retries{CustomMessage: make(map[string]int)}
	if clt.retries.CustomMessage != nil {
		for message := range clt.retries.CustomMessage {
			currentRetries.CustomMessage[message] = 0
		}
	}

	return clt.doRequestWithRetries(currentRetries, method, requestURL.String(), headers, request)
}

func (clt *Client) isLoggedIn() bool {
	return clt.accessToken != ""
}
