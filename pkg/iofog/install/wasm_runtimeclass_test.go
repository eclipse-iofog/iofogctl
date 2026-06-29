package install

import (
	"strings"
	"testing"

	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install/wasm"
)

func TestDeployWasmRuntimeClassesDeploysEachHandler(t *testing.T) {
	cfg := EdgeletInstallConfig{
		HostOS:          "linux",
		ContainerEngine: "edgelet",
		DeploymentType:  "native",
		Wasm: map[string]wasm.Pack{
			"spin":             {},
			"edgelet-wasmtime": {},
		},
	}

	var deployed []string
	err := deployWasmRuntimeClasses(
		"edge-1",
		cfg,
		func(data []byte, prefix string) (string, func(), error) {
			if !strings.HasPrefix(prefix, "runtimeclass-") {
				t.Fatalf("unexpected manifest prefix %q", prefix)
			}
			return "/tmp/" + prefix + ".yaml", func() {}, nil
		},
		func(path string) error {
			deployed = append(deployed, path)
			return nil
		},
	)
	if err != nil {
		t.Fatalf("deployWasmRuntimeClasses: %v", err)
	}
	if len(deployed) != 2 {
		t.Fatalf("deployed = %v, want 2 manifests", deployed)
	}
	if deployed[0] != "/tmp/runtimeclass-edgelet-wasmtime.yaml" {
		t.Fatalf("first deploy path = %q", deployed[0])
	}
	if deployed[1] != "/tmp/runtimeclass-spin.yaml" {
		t.Fatalf("second deploy path = %q", deployed[1])
	}
}

func TestDeployWasmRuntimeClassesSkipsWhenScopeGateClosed(t *testing.T) {
	cfg := EdgeletInstallConfig{
		HostOS:          "linux",
		ContainerEngine: "edgelet",
		DeploymentType:  "container",
		Wasm: map[string]wasm.Pack{
			"spin": {},
		},
	}

	called := false
	err := deployWasmRuntimeClasses(
		"edge-1",
		cfg,
		func([]byte, string) (string, func(), error) {
			called = true
			return "", func() {}, nil
		},
		func(string) error {
			called = true
			return nil
		},
	)
	if err != nil {
		t.Fatalf("deployWasmRuntimeClasses: %v", err)
	}
	if called {
		t.Fatal("expected RuntimeClass deploy to be skipped for container deploymentType")
	}
}

func TestPostInstallCommandsSplitBeforeBundled(t *testing.T) {
	procs, err := newDefaultEdgeletProcedures(EdgeletScriptStageDir, EdgeletInstallConfig{})
	if err != nil {
		t.Fatalf("newDefaultEdgeletProcedures: %v", err)
	}

	cfg := EdgeletInstallConfig{HostOS: "linux", ContainerEngine: "edgelet", DeploymentType: "native"}
	before := procs.postInstallCommandsBeforeBundled("edge-node", cfg, true)
	if len(before) != 4 {
		t.Fatalf("expected 4 pre-bundled post-install commands, got %d", len(before))
	}
	if !strings.Contains(before[3].cmd, "wait_edgelet_ready.sh") {
		t.Fatalf("expected wait_edgelet_ready as last pre-bundled step, got %q", before[3].cmd)
	}

	bundled := procs.postInstallBundledCommand("edge-node", cfg, true)
	if !strings.Contains(bundled.cmd, "bundled.sh") {
		t.Fatalf("expected bundled.sh command, got %q", bundled.cmd)
	}

	all := procs.postInstallCommands("edge-node", cfg, true)
	if len(all) != 5 {
		t.Fatalf("expected 5 post-install commands, got %d", len(all))
	}
}
