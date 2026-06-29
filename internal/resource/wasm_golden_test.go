package resource

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func wasmFixturePath(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata", "wasm", name)
}

func loadWasmFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(wasmFixturePath(name))
	require.NoError(t, err)
	return data
}

func TestGoldenUnmarshalWasmNativeEdgeletPackage(t *testing.T) {
	raw := loadWasmFixture(t, "native-edgelet-package.yaml")
	var doc struct {
		Package Package `yaml:"package"`
	}
	require.NoError(t, yaml.UnmarshalStrict(raw, &doc))
	pkg := doc.Package
	require.Equal(t, "1.0.0-rc.6", pkg.Version)
	require.Len(t, pkg.Wasm, 2)
	require.Contains(t, pkg.Wasm["spin"].URL, "containerd-shim-spin-v2-linux-amd64.tar.gz")
	require.Contains(t, pkg.Wasm["edgelet-wasmtime"].URL, "containerd-shim-edgelet-wasm-v2-amd64-linux-gnu.tar.gz")
}

func TestGoldenUnmarshalWasmAirgapDockerPackage(t *testing.T) {
	raw := loadWasmFixture(t, "airgap-docker-package.yaml")
	var doc struct {
		Package Package `yaml:"package"`
	}
	require.NoError(t, yaml.UnmarshalStrict(raw, &doc))
	pkg := doc.Package
	require.Len(t, pkg.Wasm, 2)
	require.Equal(t, "/media/airgap/containerd-shim-spin-v2-linux-amd64.tar.gz", pkg.Wasm["spin"].Path)
	require.Equal(t, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", pkg.Wasm["spin"].SHA256)
	require.Equal(t, "/media/airgap/containerd-shim-wasmtime-v2-linux-amd64.tar.gz", pkg.Wasm["wasmtime"].Path)
}

func TestAllowedWasmHandlersCount(t *testing.T) {
	require.Len(t, AllowedWasmHandlers, 8)
}
