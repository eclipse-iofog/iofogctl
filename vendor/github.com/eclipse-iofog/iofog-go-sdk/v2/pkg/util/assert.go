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

package util

import "fmt"

func AssertInt(in interface{}) int {
	out, ok := in.(int)
	if !ok {
		floatOut, ok := in.(float64)
		if ok {
			return int(floatOut)
		}
		panic(fmt.Sprintf(pkg.errorVariableNotInteger, in))
	}
	return out
}

func AssertBool(in interface{}) bool {
	out, ok := in.(bool)
	if !ok {
		panic(fmt.Sprintf(pkg.errorVariableNotBool, in))
	}
	return out
}
