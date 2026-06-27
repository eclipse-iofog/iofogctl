package resource

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func localFixturePath(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata", "local", name)
}

func loadLocalFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(localFixturePath(name))
	require.NoError(t, err)
	return data
}

func assertGoldenLocalControlPlane(t *testing.T, cp *LocalControlPlane) {
	t.Helper()
	require.Equal(t, "https://controller.example.com", cp.Endpoint)
	require.Equal(t, "Foo", cp.IofogUser.Name)
	require.Equal(t, "Bar", cp.IofogUser.Surname)
	require.Equal(t, email, cp.IofogUser.Email)
	require.Equal(t, "https://controller.example.com", cp.Controller.PublicUrl)
	require.Equal(t, "https://controller.example.com", cp.Controller.ConsoleUrl)
	require.Equal(t, "info", cp.Controller.LogLevel)
	require.NotNil(t, cp.Controller.Package)
	require.NotEmpty(t, cp.Controller.Package.Image)
	require.Equal(t, "embedded", cp.Auth.Mode)
	require.NotNil(t, cp.Auth.Bootstrap)
	require.Equal(t, "admin", cp.Auth.Bootstrap.Username)
	require.NotNil(t, cp.SystemAgent)
	require.NotNil(t, cp.SystemAgent.AgentConfiguration)
	require.NotNil(t, cp.SystemAgent.AgentConfiguration.Arch)
	require.Equal(t, "amd64", *cp.SystemAgent.AgentConfiguration.Arch)
	require.NotNil(t, cp.Nats)
	require.NotNil(t, cp.Nats.Enabled)
	require.True(t, *cp.Nats.Enabled)
	require.NotNil(t, cp.Events.AuditEnabled)
	require.True(t, *cp.Events.AuditEnabled)
	require.Equal(t, 14, cp.Events.RetentionDays)
}

func TestGoldenUnmarshalLocalControlPlane_Datasance(t *testing.T) {
	raw := loadLocalFixture(t, "controlplane-datasance.yaml")
	cp, err := UnmarshallLocalControlPlane(raw)
	require.NoError(t, err)
	assertGoldenLocalControlPlane(t, &cp)
}

func TestGoldenUnmarshalLocalControlPlane_Iofog(t *testing.T) {
	raw := loadLocalFixture(t, "controlplane-iofog.yaml")
	cp, err := UnmarshallLocalControlPlane(raw)
	require.NoError(t, err)
	assertGoldenLocalControlPlane(t, &cp)
}

func TestGoldenUnmarshalLocalControlPlane_WithCA(t *testing.T) {
	raw := loadLocalFixture(t, "controlplane-datasance.yaml")
	cp, err := UnmarshallLocalControlPlane(append([]byte("ca: dGVzdC1jYQ==\n"), raw...))
	require.NoError(t, err)
	require.Equal(t, "dGVzdC1jYQ==", cp.GetTrustCA())
}

func TestGoldenRoundTripLocalControlPlane(t *testing.T) {
	raw := loadLocalFixture(t, "controlplane-datasance.yaml")
	cp, err := UnmarshallLocalControlPlane(raw)
	require.NoError(t, err)

	out, err := yaml.Marshal(&cp)
	require.NoError(t, err)

	var round LocalControlPlane
	require.NoError(t, yaml.UnmarshalStrict(out, &round))
	require.Equal(t, cp.Endpoint, round.Endpoint)
	require.Equal(t, cp.Controller.PublicUrl, round.Controller.PublicUrl)
	require.Equal(t, cp.Auth.Mode, round.Auth.Mode)
	require.Equal(t, *cp.SystemAgent.AgentConfiguration.Arch, *round.SystemAgent.AgentConfiguration.Arch)
}

func requireLocalInputError(t *testing.T, err error) {
	t.Helper()
	var inputErr *util.InputError
	require.ErrorAs(t, err, &inputErr)
}

func validLocalControlPlane(t *testing.T) *LocalControlPlane {
	t.Helper()
	raw := loadLocalFixture(t, "controlplane-datasance.yaml")
	cp, err := UnmarshallLocalControlPlane(raw)
	require.NoError(t, err)
	return &cp
}

func TestValidateLocalControlPlaneMetadataRejectsControlPlaneType(t *testing.T) {
	doc := []byte(`apiVersion: datasance.com/v3
kind: LocalControlPlane
metadata:
  name: local-ecn
  controlPlaneType: local
spec:
  iofogUser:
    email: user@domain.com
  auth:
    mode: embedded
    bootstrap:
      username: admin
      password: "LocalTest12!"
  systemAgent:
    config:
      arch: amd64
`)
	err := ValidateLocalControlPlaneMetadata(doc)
	requireLocalInputError(t, err)
}

func TestValidateLocalControlPlaneMissingSystemAgent(t *testing.T) {
	cp := validLocalControlPlane(t)
	cp.SystemAgent = nil
	err := ValidateLocalControlPlane(cp)
	requireLocalInputError(t, err)
}

func TestValidateLocalControlPlaneMissingSystemAgentArch(t *testing.T) {
	cp := validLocalControlPlane(t)
	cp.SystemAgent.AgentConfiguration.Arch = nil
	err := ValidateLocalControlPlane(cp)
	requireLocalInputError(t, err)
}

func TestValidateLocalControlPlaneEndpointMismatch(t *testing.T) {
	cp := validLocalControlPlane(t)
	cp.Controller.PublicUrl = "https://other.example.com"
	err := ValidateLocalControlPlane(cp)
	requireLocalInputError(t, err)
}

func TestValidateLocalControlPlanePrivateRegistryIncomplete(t *testing.T) {
	cp := validLocalControlPlane(t)
	cp.Controller.Package = &ControllerPackage{
		Registry: "ghcr.io",
		Username: "foo",
	}
	err := ValidateLocalControlPlane(cp)
	requireLocalInputError(t, err)
}

func TestValidateLocalControlPlaneEmbeddedAuthWeakPassword(t *testing.T) {
	cp := validLocalControlPlane(t)
	cp.Auth.Bootstrap.Password = "short"
	err := ValidateLocalControlPlane(cp)
	requireLocalInputError(t, err)
}

func TestValidateLocalControlPlaneExternalAuthValid(t *testing.T) {
	cp := validLocalControlPlane(t)
	cp.Auth = Auth{
		Mode:      authModeExternal,
		IssuerUrl: "https://auth.example.com/realms/myrealm",
		Client: &AuthClient{
			ID:     "controller",
			Secret: "secret-value",
		},
	}
	require.NoError(t, ValidateLocalControlPlane(cp))
}

func TestValidateLocalControlPlaneDatabaseRequiresFields(t *testing.T) {
	cp := validLocalControlPlane(t)
	cp.Database = Database{Provider: "postgres"}
	err := ValidateLocalControlPlane(cp)
	requireLocalInputError(t, err)
}
