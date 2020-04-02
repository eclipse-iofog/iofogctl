/*
 *  *******************************************************************************
 *  * Copyright (c) 2020 Edgeworx, Inc.
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package get

import (
	"fmt"
	"testing"
)

func TestAddressConversion(t *testing.T) {
	defaultPort := "12345"
	expectedPort := "51121"
	expectedAddress := "domain.controller"
	addr, port := getAddressAndPort(fmt.Sprintf("%s:%s", expectedAddress, expectedPort), defaultPort)
	if addr != expectedAddress || port != expectedPort {
		t.Errorf("Failed Test 1 %s:%s != %s:%s", addr, port, expectedAddress, expectedPort)
	}

	// http://
	expectedPort = "61121"
	expectedAddress = "123.123.123.123"
	addr, port = getAddressAndPort(fmt.Sprintf("http://%s:%s", expectedAddress, expectedPort), defaultPort)
	if addr != expectedAddress || port != expectedPort {
		t.Errorf("Failed Test 2 %s:%s != %s:%s", addr, port, expectedAddress, expectedPort)
	}
	expectedAddress = "domain.user.com"
	addr, port = getAddressAndPort(fmt.Sprintf("http://%s:%s", expectedAddress, expectedPort), defaultPort)
	if addr != expectedAddress || port != expectedPort {
		t.Errorf("Failed Test 3 %s:%s != %s:%s", addr, port, expectedAddress, expectedPort)
	}

	// Default port
	addr, port = getAddressAndPort(fmt.Sprintf("http://%s", expectedAddress), defaultPort)
	if addr != expectedAddress || port != defaultPort {
		t.Errorf("Failed Test 4 %s:%s != %s:%s", addr, port, expectedAddress, defaultPort)
	}
	addr, port = getAddressAndPort(expectedAddress, defaultPort)
	if addr != expectedAddress || port != defaultPort {
		t.Errorf("Failed Test 5 %s:%s != %s:%s", addr, port, expectedAddress, defaultPort)
	}
}
