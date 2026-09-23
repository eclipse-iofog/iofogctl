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
	require.Contains(t, idx, config.KnowledgeKind)
	require.Contains(t, idx, config.RuntimeClassKind)
	require.Contains(t, idx, config.MicroserviceTemplateKind)
	require.Contains(t, idx, config.ApplicationKind)
	require.Contains(t, idx, config.MicroserviceKind)

	require.Less(t, idx[config.RegistryKind], idx[config.ModelKind])
	require.Less(t, idx[config.ModelKind], idx[config.KnowledgeKind])
	require.Less(t, idx[config.KnowledgeKind], idx[config.RuntimeClassKind])
	require.Less(t, idx[config.RuntimeClassKind], idx[config.MicroserviceTemplateKind])
	require.Less(t, idx[config.MicroserviceTemplateKind], idx[config.ApplicationKind])
	require.Less(t, idx[config.ApplicationKind], idx[config.MicroserviceKind])
}

func TestBuildKindHandlersIncludesV39Kinds(t *testing.T) {
	handlers := buildKindHandlers(false, 2, false, false)
	require.Contains(t, handlers, config.ModelKind)
	require.Contains(t, handlers, config.KnowledgeKind)
	require.Contains(t, handlers, config.RuntimeClassKind)
	require.Contains(t, handlers, config.MicroserviceTemplateKind)
}

func TestValidatePatchCatalogExecutorsMicroserviceOnly(t *testing.T) {
	require.NoError(t, validatePatchCatalogExecutors(map[config.Kind][]execute.Executor{
		config.MicroserviceKind: {execute.NewEmptyExecutor("app/ms")},
	}, true, false))
	require.NoError(t, validatePatchCatalogExecutors(map[config.Kind][]execute.Executor{
		config.MicroserviceKind: {execute.NewEmptyExecutor("app/ms")},
	}, false, true))
	require.NoError(t, validatePatchCatalogExecutors(map[config.Kind][]execute.Executor{
		config.MicroserviceKind: {execute.NewEmptyExecutor("app/ms")},
	}, true, true))
}

func TestValidatePatchCatalogExecutorsRejectsOtherKinds(t *testing.T) {
	err := validatePatchCatalogExecutors(map[config.Kind][]execute.Executor{
		config.MicroserviceKind: {execute.NewEmptyExecutor("app/ms")},
		config.ModelKind:        {execute.NewEmptyExecutor("llama")},
	}, true, true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Microservice-only")
	require.Contains(t, err.Error(), "--patch-model")
	require.Contains(t, err.Error(), "--patch-knowledge")
}

func TestValidatePatchCatalogExecutorsRequiresMicroservice(t *testing.T) {
	err := validatePatchCatalogExecutors(map[config.Kind][]execute.Executor{
		config.ModelKind: {execute.NewEmptyExecutor("llama")},
	}, false, true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Microservice YAML")
	require.Contains(t, err.Error(), "--patch-knowledge")
}
