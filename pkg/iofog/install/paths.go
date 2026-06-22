package install

import "strings"

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
const EdgeletScriptStageDir = "/tmp/potctl-edgelet-scripts"

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
