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
		input  string
		output string
	}{
		{
			"foo-bar", "http://foo-bar:51121",
		},
		{
			"https://foo-bar", "https://foo-bar",
		},
		{
			"http://foo-bar", "http://foo-bar",
		},
		{
			"foo-bar:1234", "http://foo-bar:1234",
		},
		{
			"1.2.3.4", "http://1.2.3.4:51121",
		},
	} {

		if result, err := GetControllerEndpoint(entry.input); result != entry.output {
			if err != nil {
				t.Errorf("Failed for input %v, when it should not", entry.input)
			} else {
				t.Errorf("Wrong result - expected: %v, actual: %v, for input: %v", entry.output, result, entry.input)
			}

		}

	}

}
