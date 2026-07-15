package install

import (
	"strings"
	"testing"

	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install/wasm"
)

func TestApplyWasmManifestRedeployRestart(t *testing.T) {
	cfg := EdgeletInstallConfig{
		HostOS:          "linux",
		ContainerEngine: "edgelet",
	}
	staged := []wasm.StagedBinary{
		{Handler: "spin", Changed: true},
	}

	cfg.SetRedeployState(true, "v1.0.0")
	if err := cfg.applyWasmManifest([]wasmManifestEntry{
		{Handler: "spin", Src: "/tmp/spin", Name: "containerd-shim-spin-v2"},
	}, staged, false, ""); err != nil {
		t.Fatalf("applyWasmManifest: %v", err)
	}
	if strings.Contains(cfg.wasmEnv, "EDGELET_WASM_SKIP_RESTART=1") {
		t.Fatalf("redeploy with changed shims should allow engine restart, got %q", cfg.wasmEnv)
	}
	if !strings.Contains(cfg.wasmEnv, "EDGELET_WASM_DEFER_ENGINE_RESTART=1") {
		t.Fatalf("redeploy with changed shims should defer engine restart to potctl, got %q", cfg.wasmEnv)
	}
	if !cfg.wasmDeferEngineRestart {
		t.Fatalf("expected wasmDeferEngineRestart for changed shims")
	}

	cfg.wasmEnv = ""
	cfg.SetRedeployState(true, "v1.0.0")
	if err := cfg.applyWasmManifest([]wasmManifestEntry{
		{Handler: "spin", Src: "/tmp/spin", Name: "containerd-shim-spin-v2"},
	}, []wasm.StagedBinary{{Handler: "spin", Changed: false}}, false, ""); err != nil {
		t.Fatalf("applyWasmManifest: %v", err)
	}
	if !strings.Contains(cfg.wasmEnv, "EDGELET_WASM_SKIP_RESTART=1") {
		t.Fatalf("redeploy with unchanged shims should skip wasm engine restart, got %q", cfg.wasmEnv)
	}

	cfg.wasmEnv = ""
	if err := cfg.applyWasmManifest([]wasmManifestEntry{
		{Handler: "spin", Src: "/tmp/spin", Name: "containerd-shim-spin-v2"},
	}, staged, true, ""); err != nil {
		t.Fatalf("applyWasmManifest: %v", err)
	}
	if !strings.Contains(cfg.wasmEnv, "EDGELET_WASM_SKIP_RESTART=1") {
		t.Fatalf("fresh install should defer wasm engine restart to start_edgelet, got %q", cfg.wasmEnv)
	}
}
