package resource

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func fixturePath(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata", "k8s", name)
}

func loadFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(fixturePath(name))
	require.NoError(t, err)
	return data
}

func assertGoldenControlPlane(t *testing.T, cp *KubernetesControlPlane) {
	t.Helper()
	require.True(t, strings.HasSuffix(cp.KubeConfig, ".kube/config"), "kubeconfig path: %s", cp.KubeConfig)
	require.Equal(t, "Foo", cp.IofogUser.Name)
	require.Equal(t, "Bar", cp.IofogUser.Surname)
	require.Equal(t, email, cp.IofogUser.Email)
	require.Equal(t, int32(2), cp.Replicas.Controller)
	require.Equal(t, int32(2), cp.Replicas.Nats)
	require.Equal(t, "https://controller.example.com", cp.Controller.PublicUrl)
	require.NotNil(t, cp.Controller.TrustProxy)
	require.True(t, *cp.Controller.TrustProxy)
	require.Equal(t, 8080, cp.Controller.ConsolePort)
	require.Equal(t, "https://controller.example.com", cp.Controller.ConsoleUrl)
	require.Equal(t, "info", cp.Controller.LogLevel)
	require.Equal(t, "embedded", cp.Auth.Mode)
	require.NotNil(t, cp.Auth.Bootstrap)
	require.Equal(t, "admin", cp.Auth.Bootstrap.Username)
	require.NotNil(t, cp.Events.AuditEnabled)
	require.True(t, *cp.Events.AuditEnabled)
	require.Equal(t, 14, cp.Events.RetentionDays)
	require.Equal(t, 86400, cp.Events.CleanupInterval)
	require.NotNil(t, cp.Events.CaptureIpAddress)
	require.True(t, *cp.Events.CaptureIpAddress)
	require.NotEmpty(t, cp.Images.Controller)
	require.NotEmpty(t, cp.Images.Router)
	require.NotEmpty(t, cp.Images.Nats)
	require.NotEmpty(t, cp.Images.Operator)
	require.NotNil(t, cp.Nats)
	require.NotNil(t, cp.Nats.Enabled)
	require.True(t, *cp.Nats.Enabled)
	require.Equal(t, "10Gi", cp.Nats.JetStream.StorageSize)
	require.Equal(t, "nginx", cp.Ingresses.Controller.IngressClassName)
	require.Equal(t, "controller.example.com", cp.Ingresses.Controller.Host)
	require.Equal(t, 5671, cp.Ingresses.Router.MessagePort)
	require.Equal(t, 4222, cp.Ingresses.Nats.ServerPort)
}

func TestGoldenUnmarshalKubernetesControlPlane_WithCA(t *testing.T) {
	raw := loadFixture(t, "controlplane-datasance.yaml")
	cp, err := UnmarshallKubernetesControlPlane(append([]byte("ca: dGVzdC1jYQ==\n"), raw...))
	require.NoError(t, err)
	require.Equal(t, "dGVzdC1jYQ==", cp.GetTrustCA())
	require.Equal(t, "dGVzdC1jYQ==", GetTrustCA(&cp))
}

func TestGoldenUnmarshalKubernetesControlPlane_Datasance(t *testing.T) {
	raw := loadFixture(t, "controlplane-datasance.yaml")
	cp, err := UnmarshallKubernetesControlPlane(raw)
	require.NoError(t, err)
	assertGoldenControlPlane(t, &cp)
}

func TestGoldenUnmarshalKubernetesControlPlane_Iofog(t *testing.T) {
	raw := loadFixture(t, "controlplane-iofog.yaml")
	cp, err := UnmarshallKubernetesControlPlane(raw)
	require.NoError(t, err)
	assertGoldenControlPlane(t, &cp)
}

func TestGoldenRoundTripKubernetesControlPlane(t *testing.T) {
	raw := loadFixture(t, "controlplane-datasance.yaml")
	cp, err := UnmarshallKubernetesControlPlane(raw)
	require.NoError(t, err)

	out, err := yaml.Marshal(&cp)
	require.NoError(t, err)

	var round KubernetesControlPlane
	require.NoError(t, yaml.UnmarshalStrict(out, &round))
	require.Equal(t, cp.KubeConfig, round.KubeConfig)
	require.Equal(t, cp.Controller.PublicUrl, round.Controller.PublicUrl)
	require.Equal(t, cp.Auth.Mode, round.Auth.Mode)
	require.Equal(t, cp.Replicas, round.Replicas)
}

func TestKubernetesControlPlane_GetTrustCA(t *testing.T) {
	cp := KubernetesControlPlane{CA: "dGVzdC1jYQ=="}
	require.Equal(t, "dGVzdC1jYQ==", cp.GetTrustCA())
	require.Equal(t, "dGVzdC1jYQ==", GetTrustCA(&cp))
}

func TestLocalControlPlane_GetTrustCA(t *testing.T) {
	cp := LocalControlPlane{CA: "dGVzdC1jYQ=="}
	require.Equal(t, "dGVzdC1jYQ==", cp.GetTrustCA())
	require.Equal(t, "dGVzdC1jYQ==", GetTrustCA(&cp))
}

func TestRemoteControlPlane_GetTrustCA(t *testing.T) {
	cp := RemoteControlPlane{CA: "dGVzdC1jYQ=="}
	require.Equal(t, "dGVzdC1jYQ==", cp.GetTrustCA())
	require.Equal(t, "dGVzdC1jYQ==", GetTrustCA(&cp))
}

func TestAuthToCPV3(t *testing.T) {
	insecure := false
	cp := KubernetesControlPlane{
		Auth: Auth{
			Mode:              "embedded",
			InsecureAllowHttp: &insecure,
			Bootstrap:         &AuthBootstrap{Username: "admin", Password: "secret"},
		},
	}
	out := AuthToCPV3(cp.Auth)
	require.Equal(t, "embedded", string(out.Mode))
	require.NotNil(t, out.Bootstrap)
	require.Equal(t, "admin", out.Bootstrap.Username)
}

func TestControllerConfigToCPV3(t *testing.T) {
	trust := true
	cfg := ControllerConfig{
		PublicUrl:  "https://controller.example.com",
		TrustProxy: &trust,
		ConsoleUrl: "https://console.example.com",
	}
	out := ControllerConfigToCPV3(cfg)
	require.Equal(t, cfg.PublicUrl, out.PublicUrl)
	require.Equal(t, cfg.ConsoleUrl, out.ConsoleUrl)
	require.True(t, *out.TrustProxy)
}
