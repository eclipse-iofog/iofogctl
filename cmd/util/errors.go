package util

import (
	"fmt"
)

// NotFound export
type NotFound struct {  
    resource    string
}
// NewNotFound export
func NewNotFound(resource string) (err *NotFound) {
    err.resource = resource
    return err
}
// Error export
func (e *NotFound) Error() string {  
    return fmt.Sprintf("Error: %s not found.", e.resource)
}

// Conflict export
type Conflict struct {
    resource string
}
// NewConflict export
func NewConflict(resource string) (err *Conflict) {
    err.resource = resource
    return err
}
// Error export
func (e *Conflict) Error() string {
    return fmt.Sprintf("Error: %s already exists.", e.resource)
}