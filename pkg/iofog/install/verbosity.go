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

package install

import (
	"fmt"
)

// Toggle HTTP output
var isVerbose bool

func SetVerbosity(verbose bool) {
	isVerbose = verbose
}

func verbose(msg string) {
	if isVerbose {
		fmt.Println(msg)
	}
}
