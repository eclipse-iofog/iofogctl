package deployairgap

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install/wasm"
)

func TestWasmRemoteStagingDir(t *testing.T) {
	got := WasmRemoteStagingDir("edge.example.com")
	wantSuffix := filepath.Join("iofogctl-airgap", "wasm", "edge-example-com")
	if !strings.Contains(got, wantSuffix) {
		t.Fatalf("WasmRemoteStagingDir = %q, want path containing %q", got, wantSuffix)
	}
}

func TestGetAirgapWasmCachePaths(t *testing.T) {
	config.Init(t.TempDir())

	cacheDir := config.GetAirgapWasmCacheDir("default", "spin", "linux", "amd64")
	if !strings.Contains(cacheDir, filepath.Join("wasm", "spin", "linux-amd64")) {
		t.Fatalf("cache dir = %q", cacheDir)
	}

	artifactPath := config.GetAirgapWasmArtifactPath("default", "spin", "linux", "amd64")
	if artifactPath != filepath.Join(cacheDir, "source") {
		t.Fatalf("artifact path = %q, want %q", artifactPath, filepath.Join(cacheDir, "source"))
	}
}

func TestEnsureWasmArtifactsFromPath(t *testing.T) {
	wasm.SetCacheRootForTest(t.TempDir())
	t.Cleanup(wasm.ResetCacheRootForTest)

	rawPath := filepath.Join(t.TempDir(), "spin-shim")
	if err := os.WriteFile(rawPath, []byte{0x7f, 'E', 'L', 'F', 0, 0, 0}, 0o755); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	staged, err := EnsureWasmArtifacts(context.Background(), "default", PlatformAMD64, map[string]rsc.WasmPack{
		"spin": {Path: rawPath},
	}, true)
	if err != nil {
		t.Fatalf("EnsureWasmArtifacts: %v", err)
	}
	if len(staged) != 1 {
		t.Fatalf("staged count = %d, want 1", len(staged))
	}
	if staged[0].CanonicalName == "" {
		t.Fatal("expected canonical shim name")
	}
	if staged[0].LocalPath == "" {
		t.Fatal("expected cached local path")
	}
}

func TestTransferWasmBinariesRequiresSSH(t *testing.T) {
	err := TransferWasmBinaries("", nil, []wasm.StagedBinary{{Handler: "spin"}})
	if err == nil {
		t.Fatal("expected error for missing host")
	}
}

func TestTransferWasmBinariesHook(t *testing.T) {
	var called bool
	SetTransferWasmBinariesHookForTest(func(host string, ssh *rsc.SSH, staged []wasm.StagedBinary) error {
		called = true
		if host != "edge.example.com" {
			t.Fatalf("host = %q", host)
		}
		if len(staged) != 1 || staged[0].Handler != "spin" {
			t.Fatalf("unexpected staged payload: %+v", staged)
		}
		staged[0].RemotePath = WasmRemoteStagingDir(host) + "/containerd-shim-spin-v2"
		return nil
	})
	t.Cleanup(ResetTransferWasmBinariesHookForTest)

	staged := []wasm.StagedBinary{{Handler: "spin", CanonicalName: "containerd-shim-spin-v2", LocalPath: "/tmp/spin"}}
	if err := TransferWasmBinaries("edge.example.com", &rsc.SSH{User: "root", KeyFile: "/tmp/key"}, staged); err != nil {
		t.Fatalf("TransferWasmBinaries: %v", err)
	}
	if !called {
		t.Fatal("expected transfer hook to run")
	}
	if staged[0].RemotePath == "" {
		t.Fatal("expected hook to populate remote path")
	}
}
