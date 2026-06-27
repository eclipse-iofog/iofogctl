package install

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteTempManifest_DesktopContainerUsesBindMountDir(t *testing.T) {
	t.Parallel()

	cfg := EdgeletInstallConfig{
		HostOS:         "darwin",
		DeploymentType: "container",
	}
	data := []byte("kind: ControlPlane\n")

	path, cleanup, err := WriteTempManifest(data, "edgelet-controlplane", cfg)
	require.NoError(t, err)
	require.NotEmpty(t, cleanup)
	t.Cleanup(cleanup)

	require.True(t, strings.HasPrefix(path, EdgeletContainerManifestDir+string(os.PathSeparator)),
		"path %q should be under %q", path, EdgeletContainerManifestDir)

	contents, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, data, contents)
}

func TestWriteTempManifest_NativeUsesSystemTemp(t *testing.T) {
	t.Parallel()

	cfg := EdgeletInstallConfig{
		HostOS:         "linux",
		DeploymentType: "native",
	}
	data := []byte("kind: ControlPlane\n")

	path, cleanup, err := WriteTempManifest(data, "edgelet-controlplane", cfg)
	require.NoError(t, err)
	t.Cleanup(cleanup)

	require.False(t, strings.HasPrefix(path, EdgeletContainerManifestDir+string(os.PathSeparator)),
		"path %q should not be under container manifest dir", path)
	require.NotEqual(t, filepath.Dir(path), EdgeletContainerManifestDir)
}

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
