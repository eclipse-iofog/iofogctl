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

import (
	"os"
	"path/filepath"
	"strings"
)

func FormatPath(input string) (string, error) {
	// Replace tilde
	if strings.Contains(input, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return strings.Replace(input, "~", homeDir, 1), nil
	}

	// Convert relative to absolute
	if strings.Contains(input, ".") {
		return filepath.Abs(input)
	}

	return input, nil
}

func Before(input string, substr string) string {
	pos := strings.Index(input, substr)
	if pos == -1 {
		return ""
	}
	return input[0:pos]
}

func After(input string, substr string) string {
	pos := strings.Index(input, substr)
	if pos == -1 || pos >= len(input)-1 {
		return ""
	}
	return input[pos+1:]
}
