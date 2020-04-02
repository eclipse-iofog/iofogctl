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

package get

import (
	"fmt"
	"testing"
)

func TestAddressConversion(t *testing.T) {
	expectedPort := "51121"
	expectedAddress := "domain.controller"
	addr, port := getAddressAndPort(fmt.Sprintf("%s:%s", expectedAddress, expectedPort), "12345")
	if addr != expectedAddress || port != expectedPort {
		t.Errorf("Failed %s:%s != %s:%s", addr, port, expectedAddress, expectedPort)
	}

	expectedPort = "61121"
	expectedAddress = "123.123.123.123"
	addr, port = getAddressAndPort(fmt.Sprintf("http://%s:%s", expectedAddress, expectedPort), "12345")
	if addr != expectedAddress || port != expectedPort {
		t.Errorf("Failed %s:%s != %s:%s", addr, port, expectedAddress, expectedPort)
	}

	expectedAddress = "domain.user.com"
	addr, port = getAddressAndPort(fmt.Sprintf("http://%s:%s", expectedAddress, expectedPort), "12345")
	if addr != expectedAddress || port != expectedPort {
		t.Errorf("Failed %s:%s != %s:%s", addr, port, expectedAddress, expectedPort)
	}
}
