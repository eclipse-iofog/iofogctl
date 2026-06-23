package install

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseEdgeletRegistryID(t *testing.T) {
	output := `ID  URL         PUBLIC  USERNAME   EMAIL
1   docker.io   true    <unknown>  <unknown>
2   from_cache  true    <unknown>  <unknown>
3   quay.io     false   john       user@domain.com
`

	id, err := ParseEdgeletRegistryID(output, "quay.io", "john")
	require.NoError(t, err)
	require.Equal(t, 3, id)

	id, err = ParseEdgeletRegistryID(output, "docker.io", "")
	require.NoError(t, err)
	require.Equal(t, 1, id)

	_, err = ParseEdgeletRegistryID(output, "missing.io", "")
	require.Error(t, err)

	_, err = ParseEdgeletRegistryID(output, "quay.io", "wrong-user")
	require.Error(t, err)
}
