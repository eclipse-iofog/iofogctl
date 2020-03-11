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

// Toggle HTTP output
var Verbose bool

func SetVerbosity(verbose bool) {
	Verbose = verbose
}

var GlobalRetriesPolicy Retries

func SetGlobalRetries(retries Retries) {
	GlobalRetriesPolicy = retries
}

type Retries struct {
	Timeout       int
	CustomMessage map[string]int
}
