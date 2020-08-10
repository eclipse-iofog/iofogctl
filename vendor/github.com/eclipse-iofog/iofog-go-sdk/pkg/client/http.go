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
	"strings"
)

func httpDo(method, url string, headers map[string]string, requestBody interface{}) (responseBody []byte, err error) {
	// Encode body
	jsonBody := ""
	if requestBody != nil {
		var jsonBodyBytes []byte
		jsonBodyBytes, err = json.Marshal(requestBody)
		if err != nil {
			return
		}
		jsonBody = string(jsonBodyBytes)
	}

	if Verbose {
		fmt.Printf("===> [%s] %s \nBody: %s\n", method, url, jsonBody)
	}

	// Instantiate request
	request, err := http.NewRequest(method, url, strings.NewReader(jsonBody))
	if err != nil {
		return
	}

	// Don't re-use connections to avoid EOF error
	request.Close = true

	// Set headers on request
	if headers != nil {
		for key, val := range headers {
			request.Header.Set(key, val)
		}
	}

	// Perform request
	client := http.Client{}
	httpResp, err := client.Do(request)
	if err != nil {
		return
	}

	// Check response
	if err = checkStatusCode(httpResp.StatusCode, method, url, httpResp.Body); err != nil {
		return
	}

	// Return body
	buf := new(bytes.Buffer)
	buf.ReadFrom(httpResp.Body)
	responseBody = buf.Bytes()
	if Verbose {
		fmt.Printf("===> Response: %s\n\n", string(responseBody))
	}
	return
}
