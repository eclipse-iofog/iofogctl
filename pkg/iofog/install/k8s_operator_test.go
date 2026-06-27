package install

import (
	"testing"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

func TestControlPlaneCRDsToInstallSingleCRD(t *testing.T) {
	crds := controlPlaneCRDsToInstall()
	require.Len(t, crds, 1)
	require.Equal(t, "controlplanes."+util.GetCliCrdGroup(), crds[0].Name)
	require.Equal(t, util.GetCliCrdGroup(), crds[0].Spec.Group)
}

func TestOperatorRBACUsesFlavorGroupNotApplications(t *testing.T) {
	rules := newOperatorRBACRules()
	group := util.GetCliCrdGroup()

	var hasControlPlaneRule bool
	for _, rule := range rules {
		for _, resource := range rule.Resources {
			require.NotContains(t, resource, "application")
		}
		for _, apiGroup := range rule.APIGroups {
			if apiGroup != group {
				continue
			}
			for _, resource := range rule.Resources {
				if resource != "controlplanes" {
					continue
				}
				hasControlPlaneRule = true
				require.ElementsMatch(t,
					[]string{"create", "delete", "get", "list", "patch", "update", "watch"},
					rule.Verbs,
				)
			}
		}
	}
	require.True(t, hasControlPlaneRule)
}

func TestSetOperatorImageDefaultsFromLdflags(t *testing.T) {
	k8s := &Kubernetes{operator: newOperatorMicroservice()}
	k8s.SetOperatorImage("")
	require.Equal(t, util.GetOperatorImage(), k8s.operator.containers[0].image)
}

func TestOperatorDeploymentWatchNamespace(t *testing.T) {
	ms := newOperatorMicroservice()
	var watchEnv *corev1.EnvVar
	for i := range ms.containers[0].env {
		if ms.containers[0].env[i].Name == "WATCH_NAMESPACE" {
			watchEnv = &ms.containers[0].env[i]
			break
		}
	}
	require.NotNil(t, watchEnv)
	require.NotNil(t, watchEnv.ValueFrom)
	require.NotNil(t, watchEnv.ValueFrom.FieldRef)
	require.Equal(t, "metadata.namespace", watchEnv.ValueFrom.FieldRef.FieldPath)
}

func TestIsSupportedControlPlaneCRD(t *testing.T) {
	expected := newControlPlaneCRD(util.GetCliCrdGroup())
	existing := expected.DeepCopy()
	require.True(t, isSupportedControlPlaneCRD(existing, expected))

	wrongGroup := newControlPlaneCRD("example.com")
	require.False(t, isSupportedControlPlaneCRD(wrongGroup, expected))
}
