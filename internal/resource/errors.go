package resource

import (
	"fmt"
	"strings"
)

type NoControlPlaneError struct {
	namespace string
	msg       string
}

func NewNoControlPlaneError(namespace string) NoControlPlaneError {
	return NoControlPlaneError{
		namespace: namespace,
		msg:       "does not contain a Control Plane",
	}
}

func (err NoControlPlaneError) Error() string {
	return fmt.Sprintf("Namespace %s %s", err.namespace, err.msg)
}

func IsNoControlPlaneError(err error) bool {
	return strings.Contains(err.Error(), NewNoControlPlaneError("").msg)
}
