package resource

import (
	"fmt"
)

// TrustProvider exposes the CLI-only spec.ca trust certificate (base64 PEM).
// Implemented by all ControlPlane kinds that support spec.ca in reference YAML.
type TrustProvider interface {
	GetTrustCA() string
}

// GetTrustCA returns spec.ca from a control plane when supported, or empty string.
func GetTrustCA(cp ControlPlane) string {
	if tp, ok := cp.(TrustProvider); ok {
		return tp.GetTrustCA()
	}
	return ""
}

// SetTrustCA sets spec.ca on supported control plane kinds.
func SetTrustCA(cp ControlPlane, caBase64 string) error {
	switch c := cp.(type) {
	case *KubernetesControlPlane:
		c.CA = caBase64
	case *LocalControlPlane:
		c.CA = caBase64
	case *RemoteControlPlane:
		c.CA = caBase64
	default:
		return fmt.Errorf("control plane does not support spec.ca")
	}
	return nil
}
