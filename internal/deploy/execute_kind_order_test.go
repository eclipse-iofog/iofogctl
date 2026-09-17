package deploy

import (
	"testing"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/stretchr/testify/require"
)

func TestKindOrderPlacesV39KindsAfterRegistryBeforeApplication(t *testing.T) {
	idx := map[config.Kind]int{}
	for i, k := range kindOrder {
		idx[k] = i
	}
	require.Contains(t, idx, config.RegistryKind)
	require.Contains(t, idx, config.ModelKind)
	require.Contains(t, idx, config.RuntimeClassKind)
	require.Contains(t, idx, config.MicroserviceTemplateKind)
	require.Contains(t, idx, config.ApplicationKind)
	require.Contains(t, idx, config.MicroserviceKind)

	require.Less(t, idx[config.RegistryKind], idx[config.ModelKind])
	require.Less(t, idx[config.ModelKind], idx[config.RuntimeClassKind])
	require.Less(t, idx[config.RuntimeClassKind], idx[config.MicroserviceTemplateKind])
	require.Less(t, idx[config.MicroserviceTemplateKind], idx[config.ApplicationKind])
	require.Less(t, idx[config.ApplicationKind], idx[config.MicroserviceKind])
}

func TestBuildKindHandlersIncludesV39Kinds(t *testing.T) {
	handlers := buildKindHandlers(false, 2, false)
	require.Contains(t, handlers, config.ModelKind)
	require.Contains(t, handlers, config.RuntimeClassKind)
	require.Contains(t, handlers, config.MicroserviceTemplateKind)
}

func TestValidatePatchModelExecutorsMicroserviceOnly(t *testing.T) {
	require.NoError(t, validatePatchModelExecutors(map[config.Kind][]execute.Executor{
		config.MicroserviceKind: {execute.NewEmptyExecutor("app/ms")},
	}))
}

func TestValidatePatchModelExecutorsRejectsOtherKinds(t *testing.T) {
	err := validatePatchModelExecutors(map[config.Kind][]execute.Executor{
		config.MicroserviceKind: {execute.NewEmptyExecutor("app/ms")},
		config.ModelKind:        {execute.NewEmptyExecutor("llama")},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "Microservice-only")
}

func TestValidatePatchModelExecutorsRequiresMicroservice(t *testing.T) {
	err := validatePatchModelExecutors(map[config.Kind][]execute.Executor{
		config.ModelKind: {execute.NewEmptyExecutor("llama")},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "Microservice YAML")
}
