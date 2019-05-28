package util

import (
	"fmt"
)

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
	return fmt.Sprintf("[ERROR] Unknown resource requested\n%s not found.", err.resource)
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
	return fmt.Sprintf("[ERROR] Resource conflict\n%s already exists.", err.resource)
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
	return "[ERROR] User Input\n" + err.message
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
	return "[ERROR] Unexpected internal behaviour\n" + err.message
}
