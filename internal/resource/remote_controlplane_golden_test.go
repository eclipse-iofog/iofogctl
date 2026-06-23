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

func remoteFixturePath(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata", "remote", name)
}

func loadRemoteFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(remoteFixturePath(name))
	require.NoError(t, err)
	return data
}

func assertGoldenRemoteControlPlane(t *testing.T, cp *RemoteControlPlane) {
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
	require.Len(t, cp.Controllers, 2)
	require.Equal(t, "remote-1", cp.Controllers[0].Name)
	require.Equal(t, "10.0.128.192", cp.Controllers[0].Host)
	require.NotNil(t, cp.Controllers[0].SystemAgent)
	require.Equal(t, "foo", cp.Controllers[0].SSH.User)
	require.Equal(t, "remote-2", cp.Controllers[1].Name)
	require.NotNil(t, cp.Controllers[1].SystemAgent)
	require.NotNil(t, cp.Nats)
	require.NotNil(t, cp.Nats.Enabled)
	require.True(t, *cp.Nats.Enabled)
	require.NotNil(t, cp.Events.AuditEnabled)
	require.True(t, *cp.Events.AuditEnabled)
	require.Equal(t, 14, cp.Events.RetentionDays)
}

func TestGoldenUnmarshalRemoteControlPlane_Datasance(t *testing.T) {
	raw := loadRemoteFixture(t, "controlplane-datasance.yaml")
	cp, err := UnmarshallRemoteControlPlane(raw)
	require.NoError(t, err)
	assertGoldenRemoteControlPlane(t, &cp)
}

func TestGoldenUnmarshalRemoteControlPlane_Iofog(t *testing.T) {
	raw := loadRemoteFixture(t, "controlplane-iofog.yaml")
	cp, err := UnmarshallRemoteControlPlane(raw)
	require.NoError(t, err)
	assertGoldenRemoteControlPlane(t, &cp)
}

func TestGoldenUnmarshalRemoteControlPlane_WithCA(t *testing.T) {
	raw := loadRemoteFixture(t, "controlplane-datasance.yaml")
	cp, err := UnmarshallRemoteControlPlane(append([]byte("ca: dGVzdC1jYQ==\n"), raw...))
	require.NoError(t, err)
	require.Equal(t, "dGVzdC1jYQ==", cp.GetTrustCA())
}

func TestGoldenRoundTripRemoteControlPlane(t *testing.T) {
	raw := loadRemoteFixture(t, "controlplane-datasance.yaml")
	cp, err := UnmarshallRemoteControlPlane(raw)
	require.NoError(t, err)

	out, err := yaml.Marshal(&cp)
	require.NoError(t, err)

	var round RemoteControlPlane
	require.NoError(t, yaml.UnmarshalStrict(out, &round))
	require.Equal(t, cp.Endpoint, round.Endpoint)
	require.Equal(t, cp.Controller.PublicUrl, round.Controller.PublicUrl)
	require.Equal(t, cp.Auth.Mode, round.Auth.Mode)
	require.Len(t, round.Controllers, len(cp.Controllers))
}

func requireRemoteInputError(t *testing.T, err error) {
	t.Helper()
	var inputErr *util.InputError
	require.ErrorAs(t, err, &inputErr)
}

func validRemoteControlPlane(t *testing.T) *RemoteControlPlane {
	t.Helper()
	raw := loadRemoteFixture(t, "controlplane-datasance.yaml")
	cp, err := UnmarshallRemoteControlPlane(raw)
	require.NoError(t, err)
	return &cp
}

func TestValidateRemoteControlPlaneMetadataRejectsControlPlaneType(t *testing.T) {
	doc := []byte(`apiVersion: datasance.com/v3
kind: ControlPlane
metadata:
  name: remote-ecn
  controlPlaneType: remote
spec:
  iofogUser:
    email: user@domain.com
  auth:
    mode: embedded
    bootstrap:
      username: admin
      password: "RemoteTest12!"
  controllers:
    - name: remote-1
      host: 10.0.0.1
      ssh:
        user: foo
        keyFile: ~/.ssh/id_rsa
      systemAgent: {}
`)
	err := ValidateRemoteControlPlaneMetadata(doc)
	requireRemoteInputError(t, err)
}

func TestValidateRemoteControlPlaneSingleControllerSQLiteOK(t *testing.T) {
	raw := loadRemoteFixture(t, "controlplane-single.yaml")
	cp, err := UnmarshallRemoteControlPlane(raw)
	require.NoError(t, err)
	require.Empty(t, cp.Database.Provider)
	require.Len(t, cp.Controllers, 1)
}

func TestValidateRemoteControlPlaneMultiControllerRequiresDatabase(t *testing.T) {
	cp := validRemoteControlPlane(t)
	cp.Database = Database{}
	err := ValidateRemoteControlPlane(cp)
	requireRemoteInputError(t, err)
}

func TestValidateRemoteControlPlaneMissingSystemAgent(t *testing.T) {
	cp := validRemoteControlPlane(t)
	cp.Controllers[0].SystemAgent = nil
	err := ValidateRemoteControlPlane(cp)
	requireRemoteInputError(t, err)
}

func TestValidateRemoteControlPlaneSystemAgentArchWhenConfigSet(t *testing.T) {
	cp := validRemoteControlPlane(t)
	cp.Controllers[0].SystemAgent.AgentConfiguration = &AgentConfiguration{}
	err := ValidateRemoteControlPlane(cp)
	requireRemoteInputError(t, err)
}

func TestValidateRemoteControlPlaneEndpointMismatch(t *testing.T) {
	cp := validRemoteControlPlane(t)
	cp.Controller.PublicUrl = "https://other.example.com"
	err := ValidateRemoteControlPlane(cp)
	requireRemoteInputError(t, err)
}

func TestValidateRemoteControlPlanePrivateRegistryIncomplete(t *testing.T) {
	cp := validRemoteControlPlane(t)
	cp.Controller.Package = &ControllerPackage{
		Registry: "ghcr.io",
		Username: "foo",
	}
	err := ValidateRemoteControlPlane(cp)
	requireRemoteInputError(t, err)
}

func TestValidateRemoteControlPlaneDuplicateControllerName(t *testing.T) {
	cp := validRemoteControlPlane(t)
	cp.Controllers[1].Name = cp.Controllers[0].Name
	err := ValidateRemoteControlPlane(cp)
	requireRemoteInputError(t, err)
}

func TestValidateRemoteControlPlaneAirgapRequiresArch(t *testing.T) {
	cp := validRemoteControlPlane(t)
	cp.Airgap = true
	err := ValidateRemoteControlPlane(cp)
	requireRemoteInputError(t, err)
}
