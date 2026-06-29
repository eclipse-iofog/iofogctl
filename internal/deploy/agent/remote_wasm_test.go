package deployagent

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	deployairgap "github.com/eclipse-iofog/iofogctl/internal/deploy/airgap"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install/wasm"
)

type wasmEdgeletStub struct {
	staged []wasm.StagedBinary
}

func (s *wasmEdgeletStub) Bootstrap() error { return nil }
func (s *wasmEdgeletStub) PrepareWasm(context.Context, string) error {
	return nil
}
func (s *wasmEdgeletStub) Configure(string, install.IofogUser, client.Options) (string, error) {
	return "", nil
}
func (s *wasmEdgeletStub) SetVersion(string) error        { return nil }
func (s *wasmEdgeletStub) SetContainerImage(string) error { return nil }
func (s *wasmEdgeletStub) SetAirgap(string) error         { return nil }
func (s *wasmEdgeletStub) CustomizeProcedures(string, *install.EdgeletProcedures) error {
	return nil
}
func (s *wasmEdgeletStub) SetWasmStaged(staged []wasm.StagedBinary) error {
	s.staged = staged
	return nil
}

func TestStageRemoteWasmAirgapInvokesTransferAndSetWasmStaged(t *testing.T) {
	wasm.SetCacheRootForTest(t.TempDir())
	t.Cleanup(wasm.ResetCacheRootForTest)

	rawPath := filepath.Join(t.TempDir(), "spin-shim")
	if err := os.WriteFile(rawPath, []byte{0x7f, 'E', 'L', 'F', 0, 0, 0}, 0o755); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	var transferred bool
	deployairgap.SetTransferWasmBinariesHookForTest(func(host string, ssh *rsc.SSH, staged []wasm.StagedBinary) error {
		transferred = true
		if host != "edge.example.com" {
			t.Fatalf("host = %q", host)
		}
		if len(staged) != 1 || staged[0].Handler != "spin" {
			t.Fatalf("unexpected staged payload: %+v", staged)
		}
		staged[0].RemotePath = deployairgap.WasmRemoteStagingDir(host) + "/containerd-shim-spin-v2"
		return nil
	})
	t.Cleanup(deployairgap.ResetTransferWasmBinariesHookForTest)

	arch := "amd64"
	cfg := &rsc.AgentConfiguration{Arch: &arch}
	stub := &wasmEdgeletStub{}
	err := stageRemoteWasmAirgap(context.Background(), "default", "edge.example.com", cfg, &rsc.SSH{User: "root", KeyFile: "/tmp/key"}, rsc.Package{
		Wasm: map[string]rsc.WasmPack{"spin": {Path: rawPath}},
	}, stub)
	if err != nil {
		t.Fatalf("stageRemoteWasmAirgap: %v", err)
	}
	if !transferred {
		t.Fatal("expected WASM transfer hook to run")
	}
	if len(stub.staged) != 1 || stub.staged[0].Handler != "spin" {
		t.Fatalf("SetWasmStaged payload = %+v", stub.staged)
	}
}
