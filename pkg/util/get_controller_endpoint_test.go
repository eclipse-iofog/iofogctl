/*
 *  *******************************************************************************
 *  * Copyright (c) 2020 Red Hat, Inc.
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */
package util

import (
	"testing"
)

func TestGetControllerEndpoint(t *testing.T) {

	for _, entry := range []struct {
		input    string
		output   string
		useHTTPS bool
	}{
		{
			"foo-bar", "http://foo-bar:51121", false,
		},
		{
			"https://foo-bar", "https://foo-bar", false,
		},
		{
			"http://foo-bar", "http://foo-bar", false,
		},
		{
			"foo-bar:1234", "http://foo-bar:1234", false,
		},
		{
			"1.2.3.4", "http://1.2.3.4:51121", false,
		},
		// HTTPS test cases
		{
			"foo-bar", "https://foo-bar:51121", true,
		},
		{
			"https://foo-bar", "https://foo-bar", true,
		},
		{
			"http://foo-bar", "http://foo-bar", true, // Should preserve existing scheme
		},
		{
			"foo-bar:1234", "https://foo-bar:1234", true,
		},
		{
			"1.2.3.4", "https://1.2.3.4:51121", true,
		},
	} {

		if result, err := GetControllerEndpoint(entry.input, entry.useHTTPS); result != entry.output {
			if err != nil {
				t.Errorf("Failed for input %v, when it should not", entry.input)
			} else {
				t.Errorf("Wrong result - expected: %v, actual: %v, for input: %v (useHTTPS: %v)", entry.output, result, entry.input, entry.useHTTPS)
			}

		}

	}

}
