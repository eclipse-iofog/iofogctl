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

func Log(callback func() error) {
	err := callback()
	if err != nil {
		PrintNotify(err.Error())
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
	header  string
	message string
}

// NewNotFoundError export
func NewNotFoundError(message string) *NotFoundError {
	return &NotFoundError{
		message: message,
		header:  "Unknown resource error",
	}
}

// Error export
func (err *NotFoundError) Error() string {
	return fmt.Sprintf("%s\n%s", err.header, err.message)
}

func IsNotFoundError(err error) bool {
	notFoundErr := NewNotFoundError("")
	return err != nil && strings.Contains(err.Error(), notFoundErr.header)
}

// ConflictError export
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

// NewInputError export
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

type UnsupportedAPIError struct {
	header  string
	message string
}

func NewUnsupportedAPIError(message string) *UnsupportedAPIError {
	return &UnsupportedAPIError{
		header:  "Unsupported API error",
		message: message,
	}
}

func (err *UnsupportedAPIError) Error() string {
	return fmt.Sprintf("%s\n%s", err.header, err.message)
}

func IsUnsupportedAPIError(err error) bool {
	apiErr := NewUnsupportedAPIError("")
	return err != nil && strings.Contains(err.Error(), apiErr.header)
}
