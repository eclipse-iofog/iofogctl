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

package util

import (
	"regexp"
)

func IsLocalHost(host string) bool {
	r := regexp.MustCompile(`^(http(s){0,1}:\/\/){0,1}(localhost|0\.0\.0\.0|127\.0\.0\.1)(:[0-9]+){0,1}`)
	return r.MatchString(host)
}
