package resource

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
