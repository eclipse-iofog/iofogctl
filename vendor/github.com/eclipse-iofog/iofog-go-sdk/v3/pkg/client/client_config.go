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

import "fmt"

// IsVerbose will Toggle HTTP output
var IsVerbose bool

func SetVerbosity(verbose bool) {
	IsVerbose = verbose
}

func Verbose(msg string) {
	if IsVerbose {
		fmt.Printf("[HTTP]: %s\n", msg)
	}
}

var GlobalRetriesPolicy Retries

func SetGlobalRetries(retries Retries) {
	GlobalRetriesPolicy = retries
}

type Retries struct {
	Timeout       int
	CustomMessage map[string]int
}
