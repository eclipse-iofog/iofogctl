package resource

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	inputvalidate "github.com/eclipse-iofog/iofogctl/internal/validate"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"gopkg.in/yaml.v2"
)

const (
	authModeEmbedded       = "embedded"
	authModeExternal       = "external"
	localControlPlaneLabel = "Local Control Plane"
)

var validDatabaseProviders = map[string]struct{}{
	"postgres": {},
	"mysql":    {},
}

var validVaultProviders = map[string]struct{}{
	"hashicorp":             {},
	"openbao":               {},
	"vault":                 {},
	"aws":                   {},
	"aws-secrets-manager":   {},
	"azure":                 {},
	"azure-key-vault":       {},
	"google":                {},
	"google-secret-manager": {},
}

// ValidateLocalControlPlaneMetadata rejects retired deploy YAML metadata fields.
func ValidateLocalControlPlaneMetadata(fullYAML []byte) error {
	var doc struct {
		Metadata map[string]interface{} `yaml:"metadata"`
	}
	if err := yaml.Unmarshal(fullYAML, &doc); err != nil {
		return util.NewUnmarshalError(err.Error())
	}
	if _, ok := doc.Metadata["controlPlaneType"]; ok {
		return util.NewInputError("metadata.controlPlaneType is retired; use kind: LocalControlPlane")
	}
	return nil
}

// ValidateLocalControlPlane validates a parsed LocalControlPlane spec.
func ValidateLocalControlPlane(cp *LocalControlPlane) error {
	if err := validateIofogUser(localControlPlaneLabel, cp.IofogUser); err != nil {
		return err
	}
	if err := validateAuth(localControlPlaneLabel, cp.Auth); err != nil {
		return err
	}
	if err := validateLocalSystemAgent(cp.SystemAgent); err != nil {
		return err
	}
	if err := validateEndpointMatch(localControlPlaneLabel, cp.Endpoint, cp.Controller.PublicUrl); err != nil {
		return err
	}
	if err := validateControllerPackage(localControlPlaneLabel, cp.Controller.Package); err != nil {
		return err
	}
	if err := validateDatabase(localControlPlaneLabel, cp.Database); err != nil {
		return err
	}
	if err := validateLocalSystemMicroservices(cp.SystemMicroservices); err != nil {
		return err
	}
	if err := validateCAField(localControlPlaneLabel, "ca", cp.CA); err != nil {
		return err
	}
	if err := validateSiteCertificateBlock(localControlPlaneLabel, "routerSiteCA", cp.RouterSiteCA); err != nil {
		return err
	}
	if err := validateSiteCertificateBlock(localControlPlaneLabel, "routerLocalCA", cp.RouterLocalCA); err != nil {
		return err
	}
	if err := validateSiteCertificateBlock(localControlPlaneLabel, "natsSiteCA", cp.NatsSiteCA); err != nil {
		return err
	}
	if err := validateSiteCertificateBlock(localControlPlaneLabel, "natsLocalCA", cp.NatsLocalCA); err != nil {
		return err
	}
	if err := validateControlPlaneTLS(localControlPlaneLabel, cp.TLS); err != nil {
		return err
	}
	if err := validateVault(localControlPlaneLabel, cp.Vault); err != nil {
		return err
	}
	return nil
}

func validateIofogUser(label string, user IofogUser) error {
	if user.Email == "" {
		return util.NewInputError(label + " iofogUser.email is required")
	}
	if rawPassword := user.GetRawPassword(); rawPassword != "" {
		if err := inputvalidate.ValidatePasswordComplexity(rawPassword); err != nil {
			return err
		}
	}
	return nil
}

func validateAuth(label string, auth Auth) error {
	switch auth.Mode {
	case authModeEmbedded:
		return validateEmbeddedAuth(label, auth)
	case authModeExternal:
		return validateExternalAuth(label, auth)
	case "":
		return util.NewInputError(label + " auth.mode is required (embedded or external)")
	default:
		return util.NewInputError(fmt.Sprintf("%s auth.mode %q is invalid (embedded or external)", label, auth.Mode))
	}
}

func validateEmbeddedAuth(label string, auth Auth) error {
	if auth.Bootstrap == nil {
		return util.NewInputError(label + " auth.bootstrap is required when auth.mode is embedded")
	}
	if auth.Bootstrap.Username == "" {
		return util.NewInputError(label + " auth.bootstrap.username is required when auth.mode is embedded")
	}
	if auth.Bootstrap.Password == "" {
		return util.NewInputError(label + " auth.bootstrap.password is required in YAML when auth.mode is embedded")
	}
	return inputvalidate.ValidatePasswordComplexity(auth.Bootstrap.Password)
}

func validateExternalAuth(label string, auth Auth) error {
	if auth.IssuerUrl == "" {
		return util.NewInputError(label + " auth.issuerUrl is required when auth.mode is external")
	}
	if auth.Client == nil || auth.Client.ID == "" {
		return util.NewInputError(label + " auth.client.id is required when auth.mode is external")
	}
	if auth.Client.Secret == "" {
		return util.NewInputError(label + " auth.client.secret is required when auth.mode is external")
	}
	return nil
}

func validateLocalSystemAgent(systemAgent *SystemAgentConfig) error {
	if systemAgent == nil {
		return util.NewInputError("Local Control Plane systemAgent is required")
	}
	if err := validateSystemAgentPackage(localControlPlaneLabel, systemAgent); err != nil {
		return err
	}
	if systemAgent.AgentConfiguration == nil || systemAgent.AgentConfiguration.Arch == nil || *systemAgent.AgentConfiguration.Arch == "" {
		return util.NewInputError("Local Control Plane systemAgent.config.arch is required")
	}
	if _, ok := ArchStringToID(*systemAgent.AgentConfiguration.Arch); !ok {
		return util.NewInputError(fmt.Sprintf("Local Control Plane systemAgent.config.arch %q is invalid", *systemAgent.AgentConfiguration.Arch))
	}
	return validateSystemAgentRouterNats(localControlPlaneLabel, systemAgent.AgentConfiguration)
}

func validateEndpointMatch(label, endpoint, publicURL string) error {
	if endpoint != "" && publicURL != "" && endpoint != publicURL {
		return util.NewInputError(label + " spec.endpoint must match spec.controller.publicUrl when both are set")
	}
	for _, value := range []string{endpoint, publicURL} {
		if value == "" {
			continue
		}
		if err := validateOptionalURL(value); err != nil {
			return util.NewInputError(fmt.Sprintf("%s endpoint URL %q is invalid: %v", label, value, err))
		}
	}
	return nil
}

func validateOptionalURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if parsed.Host == "" {
		return fmt.Errorf("missing host")
	}
	return nil
}

func validateControllerPackage(label string, pkg *ControllerPackage) error {
	if pkg == nil {
		return nil
	}
	hasRegistry := pkg.Registry != ""
	hasUsername := pkg.Username != ""
	hasPassword := pkg.Password != ""
	if hasRegistry || hasUsername || hasPassword {
		if !hasRegistry || !hasUsername || !hasPassword {
			return util.NewInputError(label + " controller.package requires registry, username, and password for private registry access")
		}
	}
	return nil
}

func validateDatabase(label string, db Database) error {
	if db.Provider == "" {
		return nil
	}
	if _, ok := validDatabaseProviders[db.Provider]; !ok {
		return util.NewInputError(fmt.Sprintf("%s database.provider %q is invalid (postgres or mysql)", label, db.Provider))
	}
	if db.User == "" || db.Host == "" || db.DatabaseName == "" || db.Password == "" || db.Port == 0 {
		return util.NewInputError(label + " database requires user, host, port, password, and databaseName when provider is set")
	}
	return nil
}

func validateLocalSystemMicroservices(sys install.RemoteSystemMicroservices) error {
	return nil
}

func validateCAField(label, name, value string) error {
	if value == "" {
		return nil
	}
	if _, err := base64.StdEncoding.DecodeString(value); err != nil {
		return util.NewInputError(fmt.Sprintf("%s %s must be valid base64", label, name))
	}
	return nil
}

func validateSiteCertificateBlock(label, name string, cert *SiteCertificate) error {
	if cert == nil {
		return nil
	}
	if cert.TLSCert != "" {
		if _, err := base64.StdEncoding.DecodeString(cert.TLSCert); err != nil {
			return util.NewInputError(fmt.Sprintf("%s %s.tlsCert must be valid base64", label, name))
		}
	}
	if cert.TLSKey != "" {
		if _, err := base64.StdEncoding.DecodeString(cert.TLSKey); err != nil {
			return util.NewInputError(fmt.Sprintf("%s %s.tlsKey must be valid base64", label, name))
		}
	}
	return nil
}

func validateControlPlaneTLS(label string, tls *ControlPlaneTLS) error {
	if tls == nil {
		return nil
	}
	for _, value := range []struct {
		name string
		raw  string
	}{
		{"tls.ca", tls.CA},
		{"tls.cert", tls.Cert},
		{"tls.key", tls.Key},
	} {
		if value.raw == "" {
			continue
		}
		if _, err := base64.StdEncoding.DecodeString(value.raw); err != nil {
			return util.NewInputError(fmt.Sprintf("%s %s must be valid base64", label, value.name))
		}
	}
	if (tls.Cert == "") != (tls.Key == "") {
		return util.NewInputError(label + " tls.cert and tls.key must both be set when either is provided")
	}
	return nil
}

func validateVault(label string, vault *VaultSpec) error {
	if vault == nil || vault.Enabled == nil || !*vault.Enabled {
		return nil
	}
	if vault.Provider == "" {
		return util.NewInputError(label + " vault.provider is required when vault.enabled is true")
	}
	if vault.BasePath == "" {
		return util.NewInputError(label + " vault.basePath is required when vault.enabled is true")
	}
	if _, ok := validVaultProviders[strings.ToLower(vault.Provider)]; !ok {
		return util.NewInputError(fmt.Sprintf("%s vault.provider %q is invalid", label, vault.Provider))
	}
	switch strings.ToLower(vault.Provider) {
	case "hashicorp", "openbao", "vault":
		if vault.Hashicorp == nil || vault.Hashicorp.Address == "" || vault.Hashicorp.Token == "" {
			return util.NewInputError(label + " vault.hashicorp is required for the selected vault provider")
		}
	case "aws", "aws-secrets-manager":
		if vault.Aws == nil || vault.Aws.Region == "" {
			return util.NewInputError(label + " vault.aws is required for the selected vault provider")
		}
	case "azure", "azure-key-vault":
		if vault.Azure == nil || vault.Azure.URL == "" {
			return util.NewInputError(label + " vault.azure is required for the selected vault provider")
		}
	case "google", "google-secret-manager":
		if vault.Google == nil || vault.Google.ProjectId == "" {
			return util.NewInputError(label + " vault.google is required for the selected vault provider")
		}
	}
	return nil
}
