package deployruntimeclass

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseRuntimeClassHandlerFromRoot(t *testing.T) {
	full := []byte(`
apiVersion: datasance.com/v3
kind: RuntimeClass
metadata:
  name: spin
handler: spin
`)
	handler, err := parseRuntimeClassHandler(full, nil)
	require.NoError(t, err)
	require.Equal(t, "spin", handler)
}

func TestParseRuntimeClassHandlerFromSpecFallback(t *testing.T) {
	spec := []byte("handler: wasmedge\n")
	handler, err := parseRuntimeClassHandler(nil, spec)
	require.NoError(t, err)
	require.Equal(t, "wasmedge", handler)
}

func TestParseRuntimeClassHandlerPrefersRootOverSpec(t *testing.T) {
	full := []byte("handler: spin\n")
	spec := []byte("handler: other\n")
	handler, err := parseRuntimeClassHandler(full, spec)
	require.NoError(t, err)
	require.Equal(t, "spin", handler)
}

func TestParseRuntimeClassHandlerMissing(t *testing.T) {
	_, err := parseRuntimeClassHandler([]byte("kind: RuntimeClass\n"), []byte("{}\n"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "handler")
}
