package deploymodel

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseModelSpecRegistryAlias(t *testing.T) {
	raw := []byte(`
repo: ai/smollm2
revision: sha256:eb3483481647229668c7625b080c2d6a632bf49535e0714198de5c026d4f41a5
registry: 1
files: []
format: gguf
`)
	spec, err := parseModelSpec(raw)
	require.NoError(t, err)
	id, err := resolveModelRegistryID(spec)
	require.NoError(t, err)
	require.Equal(t, 1, id)
	require.Equal(t, "ai/smollm2", spec.Repo)
	require.Equal(t, "gguf", spec.Format)
}

func TestParseModelSpecRegistryID(t *testing.T) {
	raw := []byte(`
repo: org/model
registryId: 2
`)
	spec, err := parseModelSpec(raw)
	require.NoError(t, err)
	id, err := resolveModelRegistryID(spec)
	require.NoError(t, err)
	require.Equal(t, 2, id)
}

func TestParseModelSpecRegistryNamedAlias(t *testing.T) {
	raw := []byte(`
repo: org/model
registry: remote
`)
	spec, err := parseModelSpec(raw)
	require.NoError(t, err)
	id, err := resolveModelRegistryID(spec)
	require.NoError(t, err)
	require.Equal(t, 1, id)
}

func TestParseModelSpecConflictingRegistry(t *testing.T) {
	raw := []byte(`
repo: org/model
registry: 1
registryId: 2
`)
	spec, err := parseModelSpec(raw)
	require.NoError(t, err)
	_, err = resolveModelRegistryID(spec)
	require.Error(t, err)
	require.Contains(t, err.Error(), "same registry")
}

func TestParseModelSpecMissingRegistry(t *testing.T) {
	raw := []byte(`
repo: org/model
`)
	spec, err := parseModelSpec(raw)
	require.NoError(t, err)
	_, err = resolveModelRegistryID(spec)
	require.Error(t, err)
	require.Contains(t, err.Error(), "registry")
}

func TestValidateModelSpecRequiresRepo(t *testing.T) {
	err := validateModelSpec(modelSpec{RegistryID: 1})
	require.Error(t, err)
	require.Contains(t, err.Error(), "repo")
}
