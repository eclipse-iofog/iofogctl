package install

import (
	"path/filepath"
	"strings"
)

// EdgeletShareDir returns the canonical on-host script directory for an edgelet host OS.
func EdgeletShareDir(hostOS string) string {
	switch normalizeEdgeletShareHostOS(hostOS) {
	case "darwin":
		return "/usr/local/share/edgelet"
	case "windows":
		return `%ProgramData%\Edgelet\scripts`
	default:
		return "/usr/share/edgelet"
	}
}

// EdgeletScriptStageDir is the transient directory used before publishing to EdgeletShareDir.
const EdgeletScriptStageDir = "/tmp/edgelet-scripts"

const wasmRemoteStageSubdir = "wasm"

// WasmLocalStageDir returns an operator-local WASM staging directory isolated per agent.
// Concurrent remote deploys must not share a single staging tree.
func WasmLocalStageDir(agentName string) string {
	return filepath.Join(EdgeletScriptStageDir, wasmRemoteStageSubdir, sanitizeWasmStageKey(agentName))
}

func sanitizeWasmStageKey(value string) string {
	if value == "" {
		return "local"
	}
	var b strings.Builder
	for _, r := range strings.ToLower(value) {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
		case r == '-' || r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	result := strings.Trim(b.String(), "-")
	if result == "" {
		return "agent"
	}
	return result
}

// EdgeletContainerManifestDir is bind-mounted into the edgelet container on desktop deploys.
const EdgeletContainerManifestDir = "/tmp/edgelet"

func normalizeEdgeletShareHostOS(hostOS string) string {
	switch strings.ToLower(strings.TrimSpace(hostOS)) {
	case "darwin", "macos", "osx":
		return "darwin"
	case "windows", "windows_nt", "win32":
		return "windows"
	default:
		return "linux"
	}
}
