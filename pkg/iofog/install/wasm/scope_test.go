package wasm

import (
	"testing"
)

func TestShouldInstallWasm(t *testing.T) {
	t.Parallel()

	base := InstallScope{
		HostOS:          "linux",
		DeploymentType:  "native",
		ContainerEngine: "edgelet",
		HasWasm:         true,
	}
	if !ShouldInstallWasm(base) {
		t.Fatal("expected wasm install for linux native edgelet")
	}

	empty := base
	empty.HasWasm = false
	if ShouldInstallWasm(empty) {
		t.Fatal("expected skip when wasm map empty")
	}

	container := base
	container.DeploymentType = "container"
	if ShouldInstallWasm(container) {
		t.Fatal("expected skip for container deployment")
	}

	podman := base
	podman.ContainerEngine = "podman"
	if ShouldInstallWasm(podman) {
		t.Fatal("expected skip for podman")
	}

	darwin := base
	darwin.HostOS = "darwin"
	if ShouldInstallWasm(darwin) {
		t.Fatal("expected skip for darwin")
	}
}

func TestNeedsEngineRestart(t *testing.T) {
	t.Parallel()

	if NeedsEngineRestart(nil, "edgelet") {
		t.Fatal("expected no restart for empty staged list")
	}
	if NeedsEngineRestart([]StagedBinary{{Changed: false}}, "docker") {
		t.Fatal("expected no restart when unchanged")
	}
	if !NeedsEngineRestart([]StagedBinary{{Changed: true}}, "edgelet") {
		t.Fatal("expected restart when shim changed")
	}
	if NeedsEngineRestart([]StagedBinary{{Changed: true}}, "podman") {
		t.Fatal("expected no restart for unsupported engine")
	}
}
