package wasm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestRuntimeClassManifest(t *testing.T) {
	t.Parallel()

	data, err := RuntimeClassManifest("spin")
	require.NoError(t, err)

	var doc runtimeClassManifest
	require.NoError(t, yaml.Unmarshal(data, &doc))
	require.Equal(t, runtimeClassAPIVersion, doc.APIVersion)
	require.Equal(t, "RuntimeClass", doc.Kind)
	require.Equal(t, "spin", doc.Metadata.Name)
	require.Equal(t, "spin", doc.Handler)
	require.NotContains(t, string(data), "---")
}

func TestRuntimeClassManifestEdgeletWasmtime(t *testing.T) {
	t.Parallel()

	data, err := RuntimeClassManifest("edgelet-wasmtime")
	require.NoError(t, err)
	require.True(t, strings.Contains(string(data), "name: edgelet-wasmtime"))
	require.True(t, strings.Contains(string(data), "handler: edgelet-wasmtime"))
}

func TestRuntimeClassManifestRequiresHandler(t *testing.T) {
	t.Parallel()

	_, err := RuntimeClassManifest("  ")
	require.Error(t, err)
}

func TestConfiguredHandlersSorted(t *testing.T) {
	t.Parallel()

	got := ConfiguredHandlers(map[string]Pack{
		"spin":             {},
		"edgelet-wasmtime": {},
	})
	require.Equal(t, []string{"edgelet-wasmtime", "spin"}, got)
}
