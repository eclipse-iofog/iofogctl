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
)

// Check error and exit
func Check(err error) {
	if err != nil {
		PrintError(err.Error())
		os.Exit(1)
	}
}

// NotFoundError export
type NotFoundError struct {
	resource string
}

// NewNotFoundError export
func NewNotFoundError(resource string) (err *NotFoundError) {
	err = new(NotFoundError)
	err.resource = resource
	return err
}

// Error export
func (err *NotFoundError) Error() string {
	return fmt.Sprintf("Unknown resource error\n%s not found.", err.resource)
}

//ConflictError export
type ConflictError struct {
	resource string
}

// NewConflictError export
func NewConflictError(resource string) (err *ConflictError) {
	err = new(ConflictError)
	err.resource = resource
	return err
}

// Error export
func (err *ConflictError) Error() string {
	return fmt.Sprintf("Resource conflict error\n%s already exists.", err.resource)
}

// InputError export
type InputError struct {
	message string
}

//NewInputError export
func NewInputError(message string) (err *InputError) {
	err = new(InputError)
	err.message = message
	return err
}

// Error export
func (err *InputError) Error() string {
	return "User input error\n" + err.message
}

// InternalError export
type InternalError struct {
	message string
}

// NewInternalError export
func NewInternalError(message string) (err *InternalError) {
	err = new(InternalError)
	err.message = message
	return err
}

// Error export
func (err *InternalError) Error() string {
	return "Unexpected internal behaviour\n" + err.message
}
