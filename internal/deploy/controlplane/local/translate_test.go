package deploylocalcontrolplane

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
	testControllerImage = "ghcr.io/datasance/controller:3.8.0-rc.5"
	testRouterImage     = "ghcr.io/datasance/router:3.8.0-rc.1"
	testNatsImage       = "ghcr.io/datasance/nats:2.14.2-rc.2"

	testIofogControllerImage = "ghcr.io/eclipse-iofog/controller:3.8.0-rc.1"
	testIofogRouterImage     = "ghcr.io/eclipse-iofog/router:3.8.0-rc.1"
	testIofogNatsImage       = "ghcr.io/eclipse-iofog/nats:2.14.2-rc.2"
)

func resourceFixturePath(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "resource", "testdata", "local", name)
}

func localFixturePath(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata", name)
}

func loadLocalControlPlaneFixture(t *testing.T, name string) rsc.LocalControlPlane {
	t.Helper()
	raw, err := os.ReadFile(resourceFixturePath(name))
	require.NoError(t, err)
	cp, err := rsc.UnmarshallLocalControlPlane(raw)
	require.NoError(t, err)
	return cp
}

func loadExpectedControlPlane(t *testing.T, name string) edgeletControlPlaneManifest {
	t.Helper()
	raw, err := os.ReadFile(localFixturePath(name))
	require.NoError(t, err)
	var manifest edgeletControlPlaneManifest
	require.NoError(t, yaml.UnmarshalStrict(raw, &manifest))
	return manifest
}

func loadExpectedRegistry(t *testing.T, name string) edgeletRegistryManifest {
	t.Helper()
	raw, err := os.ReadFile(localFixturePath(name))
	require.NoError(t, err)
	var manifest edgeletRegistryManifest
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
	cp := loadLocalControlPlaneFixture(t, "controlplane-datasance.yaml")
	opts := datasanceTranslateOptions()
	opts.RegistryID = ResolveEdgeletRegistryID(&cp, nil)
	got := TranslateEdgeletControlPlaneManifest(&cp, opts)
	want := loadExpectedControlPlane(t, "edgelet-cp-datasance.yaml")
	assertTranslatedControlPlane(t, got, want)
}

func TestTranslateEdgeletControlPlane_IofogGolden(t *testing.T) {
	cp := loadLocalControlPlaneFixture(t, "controlplane-iofog.yaml")
	opts := iofogTranslateOptions()
	opts.RegistryID = ResolveEdgeletRegistryID(&cp, nil)
	got := TranslateEdgeletControlPlaneManifest(&cp, opts)
	want := loadExpectedControlPlane(t, "edgelet-cp-iofog.yaml")
	assertTranslatedControlPlane(t, got, want)
}

func TestTranslateEdgeletRegistry_PrivateGolden(t *testing.T) {
	cp := loadLocalControlPlaneFixture(t, "controlplane-private-registry.yaml")
	require.True(t, NeedsPrivateEdgeletRegistry(&cp))
	got, err := TranslateEdgeletRegistryManifest(&cp)
	require.NoError(t, err)
	want := loadExpectedRegistry(t, "edgelet-registry-private.yaml")
	require.Equal(t, want, got)
}

func TestTranslateLocalControlPlane_WithRegistryID(t *testing.T) {
	cp := loadLocalControlPlaneFixture(t, "controlplane-private-registry.yaml")
	registryID := 3
	opts := datasanceTranslateOptions()
	opts.RegistryID = &registryID

	got := TranslateEdgeletControlPlaneManifest(&cp, opts)
	require.NotNil(t, got.Spec.Controller.Registry)
	require.Equal(t, 3, *got.Spec.Controller.Registry)

	result, err := TranslateLocalControlPlane(&cp, opts)
	require.NoError(t, err)
	require.NotEmpty(t, result.Registry)
	require.NotEmpty(t, result.ControlPlane)
}

func TestTranslateEdgeletControlPlane_DefaultControllerImage(t *testing.T) {
	cp := loadLocalControlPlaneFixture(t, "controlplane-datasance.yaml")
	cp.Controller.Package = nil
	got := TranslateEdgeletControlPlaneManifest(&cp, datasanceTranslateOptions())
	require.Equal(t, testControllerImage, got.Spec.Controller.Image)
}

func TestTranslateEdgeletControlPlane_StripsCLIOnlyFields(t *testing.T) {
	cp := loadLocalControlPlaneFixture(t, "controlplane-datasance.yaml")
	result, err := TranslateLocalControlPlane(&cp, datasanceTranslateOptions())
	require.NoError(t, err)
	require.Nil(t, result.Registry)
	body := string(result.ControlPlane)
	require.NotContains(t, body, "iofogUser")
	require.NotContains(t, body, "systemAgent")
	require.NotContains(t, body, "endpoint:")
}

func TestTranslateEdgeletControlPlane_LogLevelTopLevel(t *testing.T) {
	cp := loadLocalControlPlaneFixture(t, "controlplane-datasance.yaml")
	got := TranslateEdgeletControlPlaneManifest(&cp, datasanceTranslateOptions())
	require.Equal(t, "info", got.Spec.LogLevel)
	require.NotEmpty(t, got.Spec.Controller.Image)
}

func TestTranslateEdgeletControlPlane_ConsoleReshape(t *testing.T) {
	cp := loadLocalControlPlaneFixture(t, "controlplane-datasance.yaml")
	got := TranslateEdgeletControlPlaneManifest(&cp, datasanceTranslateOptions())
	require.NotNil(t, got.Spec.Console)
	require.Equal(t, "https://controller.example.com", got.Spec.Console.URL)
}

func TestNeedsPrivateEdgeletRegistry(t *testing.T) {
	cp := loadLocalControlPlaneFixture(t, "controlplane-datasance.yaml")
	require.False(t, NeedsPrivateEdgeletRegistry(&cp))

	privateCP := loadLocalControlPlaneFixture(t, "controlplane-private-registry.yaml")
	require.True(t, NeedsPrivateEdgeletRegistry(&privateCP))
}

func TestTranslateEdgeletControlPlane_PublicURLFromEndpoint(t *testing.T) {
	cp := loadLocalControlPlaneFixture(t, "controlplane-datasance.yaml")
	cp.Controller.PublicUrl = ""
	got := TranslateEdgeletControlPlaneManifest(&cp, datasanceTranslateOptions())
	require.Equal(t, "https://controller.example.com", got.Spec.Controller.PublicURL)
}

func TestResolveEdgeletRegistryID_OnlineDefault(t *testing.T) {
	cp := loadLocalControlPlaneFixture(t, "controlplane-datasance.yaml")
	got := ResolveEdgeletRegistryID(&cp, nil)
	require.NotNil(t, got)
	require.Equal(t, EdgeletRegistryOnline, *got)
}

func TestResolveEdgeletRegistryID_AirgapDefault(t *testing.T) {
	cp := loadLocalControlPlaneFixture(t, "controlplane-datasance.yaml")
	cp.Airgap = true
	got := ResolveEdgeletRegistryID(&cp, nil)
	require.NotNil(t, got)
	require.Equal(t, EdgeletRegistryAirgap, *got)
}

func TestResolveEdgeletRegistryID_PrivateRegistry(t *testing.T) {
	cp := loadLocalControlPlaneFixture(t, "controlplane-datasance.yaml")
	privateID := 3
	got := ResolveEdgeletRegistryID(&cp, &privateID)
	require.Equal(t, 3, *got)
}

func TestTranslateEdgeletControlPlane_AirgapRegistry(t *testing.T) {
	cp := loadLocalControlPlaneFixture(t, "controlplane-datasance.yaml")
	cp.Airgap = true
	opts := datasanceTranslateOptions()
	opts.RegistryID = ResolveEdgeletRegistryID(&cp, nil)
	got := TranslateEdgeletControlPlaneManifest(&cp, opts)
	require.NotNil(t, got.Spec.Controller.Registry)
	require.Equal(t, EdgeletRegistryAirgap, *got.Spec.Controller.Registry)
}
