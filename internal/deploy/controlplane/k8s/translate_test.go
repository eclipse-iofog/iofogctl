package deployk8scontrolplane

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	cpv3 "github.com/eclipse-iofog/iofog-operator/v3/apis/controlplanes/v3"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

const testNamespace = "test-ns"

func resourceFixturePath(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "resource", "testdata", "k8s", name)
}

func crFixturePath(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata", name)
}

func loadResourceFixture(t *testing.T, name string) rsc.KubernetesControlPlane {
	t.Helper()
	raw, err := os.ReadFile(resourceFixturePath(name))
	require.NoError(t, err)
	cp, err := rsc.UnmarshallKubernetesControlPlane(raw)
	require.NoError(t, err)
	return cp
}

func loadExpectedCR(t *testing.T, name string) cpv3.ControlPlane {
	t.Helper()
	raw, err := os.ReadFile(crFixturePath(name))
	require.NoError(t, err)
	var cp cpv3.ControlPlane
	require.NoError(t, yaml.UnmarshalStrict(raw, &cp))
	return cp
}

func stripCRSecrets(cp *cpv3.ControlPlane) {
	if cp.Spec.Auth.Bootstrap != nil {
		cp.Spec.Auth.Bootstrap.Password = ""
	}
}

func assertTranslatedCR(t *testing.T, got, want cpv3.ControlPlane) {
	t.Helper()
	stripCRSecrets(&got)
	stripCRSecrets(&want)
	require.Equal(t, want.APIVersion, got.APIVersion)
	require.Equal(t, want.Kind, got.Kind)
	require.Equal(t, want.Name, got.Name)
	require.Equal(t, want.Namespace, got.Namespace)
	require.Equal(t, want.Spec, got.Spec)
}

func TestTranslateToControlPlaneCR_DatasanceGolden(t *testing.T) {
	cp := loadResourceFixture(t, "controlplane-datasance.yaml")
	got := translateToControlPlaneCR(&cp, testNamespace, translateOptions{
		apiVersion:      "datasance.com/v3",
		crName:          "pot",
		controllerImage: "ghcr.io/datasance/controller:3.8.0-rc.5",
		routerImage:     "ghcr.io/datasance/router:3.8.0-rc.1",
		natsImage:       "ghcr.io/datasance/nats:2.14.2-rc.2",
	})
	want := loadExpectedCR(t, "cp-cr-datasance.yaml")
	assertTranslatedCR(t, got, want)
}

func TestTranslateToControlPlaneCR_IofogGolden(t *testing.T) {
	cp := loadResourceFixture(t, "controlplane-iofog.yaml")
	got := translateToControlPlaneCR(&cp, testNamespace, translateOptions{
		apiVersion:      "iofog.org/v3",
		crName:          "iofog",
		controllerImage: "ghcr.io/eclipse-iofog/controller:3.8.0-rc.5",
		routerImage:     "ghcr.io/eclipse-iofog/router:3.8.0-rc.1",
		natsImage:       "ghcr.io/eclipse-iofog/nats:2.14.2-rc.2",
	})
	want := loadExpectedCR(t, "cp-cr-iofog.yaml")
	assertTranslatedCR(t, got, want)
}

func TestTranslateToControlPlaneCR_StripsOperatorImage(t *testing.T) {
	cp := loadResourceFixture(t, "controlplane-datasance.yaml")
	require.NotEmpty(t, cp.Images.Operator)
	got := translateToControlPlaneCR(&cp, testNamespace, translateOptions{
		apiVersion:      "datasance.com/v3",
		crName:          "pot",
		controllerImage: "ghcr.io/datasance/controller:3.8.0-rc.5",
		routerImage:     "ghcr.io/datasance/router:3.8.0-rc.1",
		natsImage:       "ghcr.io/datasance/nats:2.14.2-rc.2",
	})
	require.NotContains(t, got.Spec.Images.Controller, "operator")
	require.Equal(t, "ghcr.io/datasance/controller:3.8.0-rc.5", got.Spec.Images.Controller)
}

func TestTranslateToControlPlaneCR_DefaultImagesWhenOmitted(t *testing.T) {
	cp := loadResourceFixture(t, "controlplane-datasance.yaml")
	cp.Images.Controller = ""
	cp.Images.Router = ""
	cp.Images.Nats = ""
	got := translateToControlPlaneCR(&cp, testNamespace, translateOptions{
		apiVersion:      "datasance.com/v3",
		crName:          "pot",
		controllerImage: "ghcr.io/datasance/controller:3.8.0-rc.5",
		routerImage:     "ghcr.io/datasance/router:3.8.0-rc.1",
		natsImage:       "ghcr.io/datasance/nats:2.14.2-rc.2",
	})
	require.Equal(t, "ghcr.io/datasance/controller:3.8.0-rc.5", got.Spec.Images.Controller)
	require.Equal(t, "ghcr.io/datasance/router:3.8.0-rc.1", got.Spec.Images.Router)
	require.Equal(t, "ghcr.io/datasance/nats:2.14.2-rc.2", got.Spec.Images.Nats)
}

func TestTranslateToControlPlaneCR_CRNameFromLdflagDefault(t *testing.T) {
	cp := loadResourceFixture(t, "controlplane-iofog.yaml")
	got := TranslateToControlPlaneCR(&cp, testNamespace)
	require.Equal(t, util.GetCliCpCrName(), got.Name)
	require.Equal(t, util.GetCliApiVersion(), got.APIVersion)
}
