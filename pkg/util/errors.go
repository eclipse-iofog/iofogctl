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
	"strings"
)

// Check error and exit
func Check(err error) {
	if err != nil {
		PrintError(err.Error())
		os.Exit(1)
	}
}

type Error struct {
	message string
}

func NewError(message string) *Error {
	return &Error{
		message: message,
	}
}

func (err *Error) Error() string {
	return err.message
}

// NotFoundError export
type NotFoundError struct {
	message string
}

// NewNotFoundError export
func NewNotFoundError(message string) *NotFoundError {
	return &NotFoundError{
		message: message,
	}
}

// Error export
func (err *NotFoundError) Error() string {
	return fmt.Sprintf("Unknown resource error\n%s", err.message)
}

//ConflictError export
type ConflictError struct {
	message string
}

// NewConflictError export
func NewConflictError(message string) *ConflictError {
	return &ConflictError{
		message: message,
	}
}

// Error export
func (err *ConflictError) Error() string {
	return fmt.Sprintf("Resource conflict error\n%s", err.message)
}

// InputError export
type InputError struct {
	message string
}

//NewInputError export
func NewInputError(message string) *InputError {
	return &InputError{
		message: message,
	}
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
func NewInternalError(message string) *InternalError {
	return &InternalError{
		message: message,
	}
}

// Error export
func (err *InternalError) Error() string {
	return "Unexpected internal behaviour\n" + err.message
}

// HTTPError export
type HTTPError struct {
	message string
	Code    int
}

// NewHTTPError export
func NewHTTPError(message string, code int) *HTTPError {
	return &HTTPError{
		message: message,
		Code:    code,
	}
}

// Error export
func (err *HTTPError) Error() string {
	return "Unexpected HTTP response\n" + err.message
}

type UnmarshalError struct {
	message string
}

func NewUnmarshalError(message string) *UnmarshalError {
	return &UnmarshalError{
		message: message,
	}
}

func (err *UnmarshalError) Error() string {
	return fmt.Sprintf("Failed to unmarshal input file. \n%s\nMake sure to use camel case field names. E.g. `keyFile: ~/.ssh/id_rsa`", err.message)
}

type NoConfigError struct {
	message string
}

func NewNoConfigError(resource string) *NoConfigError {
	res := strings.ToLower(resource)
	var kubeText string
	if res == "connector" || res == "controller" {
		kubeText = "Kube Config and"
	}
	message := fmt.Sprintf("Cannot perform command because %s SSH details for this %s are not available. Use the configure command to add required details.", kubeText, res)

	return &NoConfigError{
		message: message,
	}
}

func (err *NoConfigError) Error() string {
	return err.message
}
