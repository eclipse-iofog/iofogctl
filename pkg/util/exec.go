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
	"bytes"
	"os"
	"os/exec"
)

// Exec command
func Exec(env, cmdName string, args ...string) (stdout bytes.Buffer, err error) {
	// Instantiate command object
	cmd := exec.Command(cmdName, args...)

	// Instantiate output objects
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Set env vars
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, env)

	// Run command
	err = cmd.Run()
	if err != nil {
		err = NewInternalError(stderr.String())
		return
	}
	return
}
