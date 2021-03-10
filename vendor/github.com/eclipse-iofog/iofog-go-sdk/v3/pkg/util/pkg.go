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

var pkg struct {
	errorVariableNotInteger string
	errorVariableNotBool    string
}

func init() {
	pkg.errorVariableNotInteger = "Variable (%s) is not of type integer"
	pkg.errorVariableNotBool = "Variable (%s) is not of type bool"
}
