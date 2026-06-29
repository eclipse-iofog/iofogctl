package wasm

import (
	"strings"
)

// InstallScope captures edgelet install fields used for WASM layer gating.
type InstallScope struct {
	HostOS          string
	DeploymentType  string
	ContainerEngine string
	HasWasm         bool
}

// ShouldInstallWasm reports whether the WASM bootstrap layer should run.
func ShouldInstallWasm(scope InstallScope) bool {
	hostOS := strings.ToLower(strings.TrimSpace(scope.HostOS))
	if hostOS == "" {
		hostOS = "linux"
	}
	if hostOS != "linux" {
		return false
	}

	deploy := strings.ToLower(strings.TrimSpace(scope.DeploymentType))
	if deploy == "" {
		deploy = "native"
	}
	if deploy == "container" {
		return false
	}

	engine := strings.ToLower(strings.TrimSpace(scope.ContainerEngine))
	if engine == "" {
		engine = "edgelet"
	}
	if engine == "podman" {
		return false
	}

	return scope.HasWasm
}

// NeedsEngineRestart reports whether the container engine should restart after shim install.
func NeedsEngineRestart(staged []StagedBinary, engine string) bool {
	engine = strings.ToLower(strings.TrimSpace(engine))
	if engine == "" {
		engine = "edgelet"
	}
	if engine != "edgelet" && engine != "docker" {
		return false
	}
	for _, item := range staged {
		if item.Changed {
			return true
		}
	}
	return false
}
