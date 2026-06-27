package deployremotecontrolplane

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

const testNamespace = "test-ns"

const (
	testControllerImage = "ghcr.io/datasance/controller:3.8.0-rc.1"
	testRouterImage     = "ghcr.io/datasance/router:3.8.0-rc.1"
	testNatsImage       = "ghcr.io/datasance/nats:2.14.2-rc.2"

	testIofogControllerImage = "ghcr.io/eclipse-iofog/controller:3.8.0-rc.1"
	testIofogRouterImage     = "ghcr.io/eclipse-iofog/router:3.8.0-rc.1"
	testIofogNatsImage       = "ghcr.io/eclipse-iofog/nats:2.14.2-rc.2"
)

func resourceFixturePath(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "resource", "testdata", "remote", name)
}

func remoteFixturePath(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata", name)
}

func loadRemoteControlPlaneFixture(t *testing.T, name string) rsc.RemoteControlPlane {
	t.Helper()
	raw, err := os.ReadFile(resourceFixturePath(name))
	require.NoError(t, err)
	cp, err := rsc.UnmarshallRemoteControlPlane(raw)
	require.NoError(t, err)
	return cp
}

func loadExpectedControlPlane(t *testing.T, name string) edgeletControlPlaneManifest {
	t.Helper()
	raw, err := os.ReadFile(remoteFixturePath(name))
	require.NoError(t, err)
	var manifest edgeletControlPlaneManifest
	require.NoError(t, yaml.UnmarshalStrict(raw, &manifest))
	return manifest
}

func stripManifestSecrets(m *edgeletControlPlaneManifest) {
	if m.Spec.Auth.Bootstrap != nil {
		m.Spec.Auth.Bootstrap.Password = ""
	}
}

func assertTranslatedControlPlane(t *testing.T, got, want edgeletControlPlaneManifest) {
	t.Helper()
	stripManifestSecrets(&got)
	stripManifestSecrets(&want)
	require.Equal(t, want, got)
}

func datasanceTranslateOptions() TranslateOptions {
	return TranslateOptions{
		Name:            "iofog",
		Namespace:       testNamespace,
		ControllerImage: testControllerImage,
		RouterImage:     testRouterImage,
		NatsImage:       testNatsImage,
	}
}

func iofogTranslateOptions() TranslateOptions {
	return TranslateOptions{
		Name:            "iofog",
		Namespace:       testNamespace,
		ControllerImage: testIofogControllerImage,
		RouterImage:     testIofogRouterImage,
		NatsImage:       testIofogNatsImage,
	}
}

func TestTranslateEdgeletControlPlane_DatasanceGolden(t *testing.T) {
	cp := loadRemoteControlPlaneFixture(t, "controlplane-datasance.yaml")
	ctrl := cp.Controllers[0]
	opts := datasanceTranslateOptions()
	opts.RegistryID = ResolveEdgeletRegistryID(&cp, nil)
	got := TranslateEdgeletControlPlaneManifest(&cp, &ctrl, opts)
	want := loadExpectedControlPlane(t, "edgelet-cp-datasance.yaml")
	assertTranslatedControlPlane(t, got, want)
}

func TestTranslateEdgeletControlPlane_IofogGolden(t *testing.T) {
	cp := loadRemoteControlPlaneFixture(t, "controlplane-iofog.yaml")
	ctrl := cp.Controllers[0]
	opts := iofogTranslateOptions()
	opts.RegistryID = ResolveEdgeletRegistryID(&cp, nil)
	got := TranslateEdgeletControlPlaneManifest(&cp, &ctrl, opts)
	want := loadExpectedControlPlane(t, "edgelet-cp-iofog.yaml")
	assertTranslatedControlPlane(t, got, want)
}

func TestTranslateEdgeletControlPlane_PerControllerTLSOverride(t *testing.T) {
	cp := loadRemoteControlPlaneFixture(t, "controlplane-datasance.yaml")
	cp.TLS = &rsc.ControlPlaneTLS{
		CA:   "global-ca",
		Cert: "global-cert",
		Key:  "global-key",
	}
	ctrl := cp.Controllers[0]
	ctrl.TLS = &rsc.ControlPlaneTLS{
		CA:   "host-ca",
		Cert: "host-cert",
		Key:  "host-key",
	}

	got := TranslateEdgeletControlPlaneManifest(&cp, &ctrl, datasanceTranslateOptions())
	require.NotNil(t, got.Spec.TLS)
	require.NotNil(t, got.Spec.TLS.Base64)
	require.Equal(t, "host-ca", got.Spec.TLS.Base64.CA)
	require.Equal(t, "host-cert", got.Spec.TLS.Base64.Cert)
	require.Equal(t, "host-key", got.Spec.TLS.Base64.Key)
}

func TestTranslateEdgeletControlPlane_StripsCLIOnlyFields(t *testing.T) {
	cp := loadRemoteControlPlaneFixture(t, "controlplane-datasance.yaml")
	ctrl := cp.Controllers[0]
	result, err := TranslateRemoteControlPlane(&cp, &ctrl, datasanceTranslateOptions())
	require.NoError(t, err)
	require.Nil(t, result.Registry)
	body := string(result.ControlPlane)
	require.NotContains(t, body, "iofogUser")
	require.NotContains(t, body, "systemAgent")
	require.NotContains(t, body, "controllers:")
}

func TestResolveEdgeletRegistryID_AirgapDefault(t *testing.T) {
	cp := loadRemoteControlPlaneFixture(t, "controlplane-datasance.yaml")
	cp.Airgap = true
	got := ResolveEdgeletRegistryID(&cp, nil)
	require.NotNil(t, got)
	require.Equal(t, EdgeletRegistryAirgap, *got)
}
