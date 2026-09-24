package deployknowledge

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseKnowledgeSpecRegistryAlias(t *testing.T) {
	raw := []byte(`
repo: acme/product-manuals
revision: 9f3c111122223333444455556666777788889999
registry: 1
files: []
format: jsonl
`)
	spec, err := parseKnowledgeSpec(raw)
	require.NoError(t, err)
	id, err := resolveKnowledgeRegistryID(spec)
	require.NoError(t, err)
	require.Equal(t, 1, id)
	require.Equal(t, "acme/product-manuals", spec.Repo)
	require.Equal(t, "jsonl", spec.Format)
}

func TestParseKnowledgeSpecRegistryID(t *testing.T) {
	raw := []byte(`
repo: org/docs
registryId: 2
`)
	spec, err := parseKnowledgeSpec(raw)
	require.NoError(t, err)
	id, err := resolveKnowledgeRegistryID(spec)
	require.NoError(t, err)
	require.Equal(t, 2, id)
}

func TestParseKnowledgeSpecRegistryNamedAlias(t *testing.T) {
	raw := []byte(`
repo: org/docs
registry: remote
`)
	spec, err := parseKnowledgeSpec(raw)
	require.NoError(t, err)
	id, err := resolveKnowledgeRegistryID(spec)
	require.NoError(t, err)
	require.Equal(t, 1, id)
}

func TestParseKnowledgeSpecConflictingRegistry(t *testing.T) {
	raw := []byte(`
repo: org/docs
registry: 1
registryId: 2
`)
	spec, err := parseKnowledgeSpec(raw)
	require.NoError(t, err)
	_, err = resolveKnowledgeRegistryID(spec)
	require.Error(t, err)
	require.Contains(t, err.Error(), "same registry")
}

func TestParseKnowledgeSpecMissingRegistry(t *testing.T) {
	raw := []byte(`
repo: org/docs
`)
	spec, err := parseKnowledgeSpec(raw)
	require.NoError(t, err)
	_, err = resolveKnowledgeRegistryID(spec)
	require.Error(t, err)
	require.Contains(t, err.Error(), "registry")
}

func TestValidateKnowledgeSpecRequiresRepo(t *testing.T) {
	err := validateKnowledgeSpec(knowledgeSpec{RegistryID: 1})
	require.Error(t, err)
	require.Contains(t, err.Error(), "repo")
}
