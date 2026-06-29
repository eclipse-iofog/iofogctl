package resource

import (
	"testing"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/stretchr/testify/require"
)

func requireWasmInputError(t *testing.T, err error, contains string) {
	t.Helper()
	require.Error(t, err)
	var inputErr *util.InputError
	require.ErrorAs(t, err, &inputErr)
	require.Contains(t, err.Error(), contains)
}

func TestValidatePackageWasmUnknownHandler(t *testing.T) {
	err := validatePackageWasm("Agent", Package{
		Wasm: map[string]WasmPack{
			"not-a-handler": {URL: "https://example.com/shim.tar.gz"},
		},
	}, strPtr("edgelet"))
	requireWasmInputError(t, err, "unknown handler")
}

func TestValidatePackageWasmMissingSource(t *testing.T) {
	err := validatePackageWasm("Agent", Package{
		Wasm: map[string]WasmPack{
			"spin": {},
		},
	}, strPtr("edgelet"))
	requireWasmInputError(t, err, "requires url or path")
}

func TestValidatePackageWasmBothURLAndPath(t *testing.T) {
	err := validatePackageWasm("Agent", Package{
		Wasm: map[string]WasmPack{
			"spin": {
				URL:  "https://example.com/shim.tar.gz",
				Path: "/media/shim.tar.gz",
			},
		},
	}, strPtr("edgelet"))
	requireWasmInputError(t, err, "either url or path, not both")
}

func TestValidatePackageWasmInvalidSHA256(t *testing.T) {
	err := validatePackageWasm("Agent", Package{
		Wasm: map[string]WasmPack{
			"spin": {
				URL:    "https://example.com/shim.tar.gz",
				SHA256: "not-hex",
			},
		},
	}, strPtr("edgelet"))
	requireWasmInputError(t, err, "sha256")
}

func TestValidatePackageWasmPodmanRejected(t *testing.T) {
	err := validatePackageWasm("Agent", Package{
		Wasm: map[string]WasmPack{
			"spin": {URL: "https://example.com/shim.tar.gz"},
		},
	}, strPtr("podman"))
	requireWasmInputError(t, err, "podman")
}

func TestValidatePackageWasmEmptyMapOK(t *testing.T) {
	require.NoError(t, validatePackageWasm("Agent", Package{}, strPtr("podman")))
}

func TestValidatePackageWasmDockerOK(t *testing.T) {
	require.NoError(t, validatePackageWasm("Agent", Package{
		Wasm: map[string]WasmPack{
			"wasmtime": {Path: "/media/shim.tar.gz"},
		},
	}, strPtr("docker")))
}

func TestValidateLocalControlPlaneSystemAgentWasmPodman(t *testing.T) {
	engine := "podman"
	cp := validLocalControlPlane(t)
	cp.SystemAgent.Package = Package{
		Wasm: map[string]WasmPack{
			"spin": {URL: "https://example.com/shim.tar.gz"},
		},
	}
	cp.SystemAgent.AgentConfiguration.ContainerEngine = &engine
	err := ValidateLocalControlPlane(cp)
	requireWasmInputError(t, err, "podman")
}

func TestValidateRemoteControlPlaneSystemAgentWasmUnknownHandler(t *testing.T) {
	cp := validRemoteControlPlane(t)
	cp.Controllers[0].SystemAgent.Package = Package{
		Wasm: map[string]WasmPack{
			"bogus": {Path: "/media/shim.tar.gz"},
		},
	}
	err := ValidateRemoteControlPlane(cp)
	requireWasmInputError(t, err, "unknown handler")
}
