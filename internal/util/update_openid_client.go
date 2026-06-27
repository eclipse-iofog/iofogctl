package util

import (
	"github.com/eclipse-iofog/iofogctl/internal/resource"
)

// UpdateECNViewerClientRootURL is retired in v3.8 (Keycloak viewer client replaced by embedded/external OIDC).
func UpdateECNViewerClientRootURL(_ resource.Auth, _ string) error {
	return nil
}
