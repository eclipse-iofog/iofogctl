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

type NoSSHConfigError struct {
	message string
}

func NewNoSSHConfigError(resource string) *NoSSHConfigError {
	message := fmt.Sprintf("Cannot perform command because SSH details for this %s are not available. Use the configure command to add required details.", resource)

	return &NoSSHConfigError{
		message: message,
	}
}

func (err *NoSSHConfigError) Error() string {
	return err.message
}

type NoKubeConfigError struct {
	message string
}

func NewNoKubeConfigError(resource string) *NoKubeConfigError {
	message := fmt.Sprintf("Cannot perform command because Kube config for this %s is not available. Use the configure command to add required details.", resource)

	return &NoKubeConfigError{
		message: message,
	}
}

func (err *NoKubeConfigError) Error() string {
	return err.message
}
