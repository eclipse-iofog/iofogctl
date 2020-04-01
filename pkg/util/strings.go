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
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func FormatPath(input string) (string, error) {
	if len(input) == 0 {
		return input, nil
	}

	// Replace tilde
	if string(input[0]) == "~" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return input, err
		}
		return strings.Replace(input, "~", homeDir, 1), nil
	}

	// Convert relative to absolute
	if string(input[0]) == "." {
		return filepath.Abs(input)
	}

	return input, nil
}

func Before(input, substr string) string {
	pos := strings.Index(input, substr)
	if pos == -1 {
		return input
	}
	return input[0:pos]
}

func After(input, substr string) string {
	pos := strings.Index(input, substr)
	if pos == -1 || pos+1 > len(input)-1 {
		return ""
	}
	return input[pos+1:]
}

func AfterLast(input, substr string) string {
	pos := strings.LastIndex(input, substr)
	if pos == -1 || pos+1 > len(input)-1 {
		return ""
	}
	return input[pos+1:]
}

func IsLowerAlphanumeric(input string) error {
	if len(input) <= 2 {
		return NewInputError(fmt.Sprintf("Resource [%s]: Invalid name. Names must be atleast 3 characters in length.", input))
	}
	regex := regexp.MustCompile("^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$")
	if !regex.MatchString(input) {
		return NewInputError(fmt.Sprintf("Resource [%s]: Invalid name. Names must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character", input))
	}
	return nil
}
