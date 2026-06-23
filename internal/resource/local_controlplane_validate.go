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
	authModeEmbedded = "embedded"
	authModeExternal = "external"
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
	if err := validateLocalIofogUser(cp.IofogUser); err != nil {
		return err
	}
	if err := validateLocalAuth(cp.Auth); err != nil {
		return err
	}
	if err := validateLocalSystemAgent(cp.SystemAgent); err != nil {
		return err
	}
	if err := validateLocalEndpoint(cp.Endpoint, cp.Controller.PublicUrl); err != nil {
		return err
	}
	if err := validateControllerPackage(cp.Controller.Package); err != nil {
		return err
	}
	if err := validateLocalDatabase(cp.Database); err != nil {
		return err
	}
	if err := validateLocalSystemMicroservices(cp.SystemMicroservices); err != nil {
		return err
	}
	if err := validateLocalCAField("ca", cp.CA); err != nil {
		return err
	}
	if err := validateSiteCertificateBlock("routerSiteCA", cp.RouterSiteCA); err != nil {
		return err
	}
	if err := validateSiteCertificateBlock("routerLocalCA", cp.RouterLocalCA); err != nil {
		return err
	}
	if err := validateSiteCertificateBlock("natsSiteCA", cp.NatsSiteCA); err != nil {
		return err
	}
	if err := validateSiteCertificateBlock("natsLocalCA", cp.NatsLocalCA); err != nil {
		return err
	}
	if err := validateControlPlaneTLS(cp.TLS); err != nil {
		return err
	}
	if err := validateLocalVault(cp.Vault); err != nil {
		return err
	}
	return nil
}

func validateLocalIofogUser(user IofogUser) error {
	if user.Email == "" {
		return util.NewInputError("Local Control Plane iofogUser.email is required")
	}
	if rawPassword := user.GetRawPassword(); rawPassword != "" {
		if err := inputvalidate.ValidatePasswordComplexity(rawPassword); err != nil {
			return err
		}
	}
	return nil
}

func validateLocalAuth(auth Auth) error {
	switch auth.Mode {
	case authModeEmbedded:
		return validateEmbeddedAuth(auth)
	case authModeExternal:
		return validateExternalAuth(auth)
	case "":
		return util.NewInputError("Local Control Plane auth.mode is required (embedded or external)")
	default:
		return util.NewInputError(fmt.Sprintf("Local Control Plane auth.mode %q is invalid (embedded or external)", auth.Mode))
	}
}

func validateEmbeddedAuth(auth Auth) error {
	if auth.Bootstrap == nil {
		return util.NewInputError("Local Control Plane auth.bootstrap is required when auth.mode is embedded")
	}
	if auth.Bootstrap.Username == "" {
		return util.NewInputError("Local Control Plane auth.bootstrap.username is required when auth.mode is embedded")
	}
	if auth.Bootstrap.Password == "" {
		return util.NewInputError("Local Control Plane auth.bootstrap.password is required in YAML when auth.mode is embedded")
	}
	return inputvalidate.ValidatePasswordComplexity(auth.Bootstrap.Password)
}

func validateExternalAuth(auth Auth) error {
	if auth.IssuerUrl == "" {
		return util.NewInputError("Local Control Plane auth.issuerUrl is required when auth.mode is external")
	}
	if auth.Client == nil || auth.Client.ID == "" {
		return util.NewInputError("Local Control Plane auth.client.id is required when auth.mode is external")
	}
	if auth.Client.Secret == "" {
		return util.NewInputError("Local Control Plane auth.client.secret is required when auth.mode is external")
	}
	return nil
}

func validateLocalSystemAgent(systemAgent *SystemAgentConfig) error {
	if systemAgent == nil {
		return util.NewInputError("Local Control Plane systemAgent is required")
	}
	if systemAgent.AgentConfiguration == nil || systemAgent.AgentConfiguration.Arch == nil || *systemAgent.AgentConfiguration.Arch == "" {
		return util.NewInputError("Local Control Plane systemAgent.config.arch is required")
	}
	if _, ok := ArchStringToID(*systemAgent.AgentConfiguration.Arch); !ok {
		return util.NewInputError(fmt.Sprintf("Local Control Plane systemAgent.config.arch %q is invalid", *systemAgent.AgentConfiguration.Arch))
	}
	return nil
}

func validateLocalEndpoint(endpoint, publicURL string) error {
	if endpoint != "" && publicURL != "" && endpoint != publicURL {
		return util.NewInputError("Local Control Plane spec.endpoint must match spec.controller.publicUrl when both are set")
	}
	for _, value := range []string{endpoint, publicURL} {
		if value == "" {
			continue
		}
		if err := validateOptionalURL(value); err != nil {
			return util.NewInputError(fmt.Sprintf("Local Control Plane endpoint URL %q is invalid: %v", value, err))
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

func validateControllerPackage(pkg *ControllerPackage) error {
	if pkg == nil {
		return nil
	}
	hasRegistry := pkg.Registry != ""
	hasUsername := pkg.Username != ""
	hasPassword := pkg.Password != ""
	if hasRegistry || hasUsername || hasPassword {
		if !hasRegistry || !hasUsername || !hasPassword {
			return util.NewInputError("Local Control Plane controller.package requires registry, username, and password for private registry access")
		}
	}
	return nil
}

func validateLocalDatabase(db Database) error {
	if db.Provider == "" {
		return nil
	}
	if _, ok := validDatabaseProviders[db.Provider]; !ok {
		return util.NewInputError(fmt.Sprintf("Local Control Plane database.provider %q is invalid (postgres or mysql)", db.Provider))
	}
	if db.User == "" || db.Host == "" || db.DatabaseName == "" || db.Password == "" || db.Port == 0 {
		return util.NewInputError("Local Control Plane database requires user, host, port, password, and databaseName when provider is set")
	}
	return nil
}

func validateLocalSystemMicroservices(sys install.RemoteSystemMicroservices) error {
	return nil
}

func validateLocalCAField(name, value string) error {
	if value == "" {
		return nil
	}
	if _, err := base64.StdEncoding.DecodeString(value); err != nil {
		return util.NewInputError(fmt.Sprintf("Local Control Plane %s must be valid base64", name))
	}
	return nil
}

func validateSiteCertificateBlock(name string, cert *SiteCertificate) error {
	if cert == nil {
		return nil
	}
	if cert.TLSCert != "" {
		if _, err := base64.StdEncoding.DecodeString(cert.TLSCert); err != nil {
			return util.NewInputError(fmt.Sprintf("Local Control Plane %s.tlsCert must be valid base64", name))
		}
	}
	if cert.TLSKey != "" {
		if _, err := base64.StdEncoding.DecodeString(cert.TLSKey); err != nil {
			return util.NewInputError(fmt.Sprintf("Local Control Plane %s.tlsKey must be valid base64", name))
		}
	}
	return nil
}

func validateControlPlaneTLS(tls *ControlPlaneTLS) error {
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
			return util.NewInputError(fmt.Sprintf("Local Control Plane %s must be valid base64", value.name))
		}
	}
	if (tls.Cert == "") != (tls.Key == "") {
		return util.NewInputError("Local Control Plane tls.cert and tls.key must both be set when either is provided")
	}
	return nil
}

func validateLocalVault(vault *VaultSpec) error {
	if vault == nil || vault.Enabled == nil || !*vault.Enabled {
		return nil
	}
	if vault.Provider == "" {
		return util.NewInputError("Local Control Plane vault.provider is required when vault.enabled is true")
	}
	if vault.BasePath == "" {
		return util.NewInputError("Local Control Plane vault.basePath is required when vault.enabled is true")
	}
	if _, ok := validVaultProviders[strings.ToLower(vault.Provider)]; !ok {
		return util.NewInputError(fmt.Sprintf("Local Control Plane vault.provider %q is invalid", vault.Provider))
	}
	switch strings.ToLower(vault.Provider) {
	case "hashicorp", "openbao", "vault":
		if vault.Hashicorp == nil || vault.Hashicorp.Address == "" || vault.Hashicorp.Token == "" {
			return util.NewInputError("Local Control Plane vault.hashicorp is required for the selected vault provider")
		}
	case "aws", "aws-secrets-manager":
		if vault.Aws == nil || vault.Aws.Region == "" {
			return util.NewInputError("Local Control Plane vault.aws is required for the selected vault provider")
		}
	case "azure", "azure-key-vault":
		if vault.Azure == nil || vault.Azure.URL == "" {
			return util.NewInputError("Local Control Plane vault.azure is required for the selected vault provider")
		}
	case "google", "google-secret-manager":
		if vault.Google == nil || vault.Google.ProjectId == "" {
			return util.NewInputError("Local Control Plane vault.google is required for the selected vault provider")
		}
	}
	return nil
}
