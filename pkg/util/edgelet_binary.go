package util

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var supportedEdgeletArches = map[string]struct{}{
	"amd64":   {},
	"arm64":   {},
	"arm":     {},
	"riscv64": {},
}

// NormalizeEdgeletOS maps host GOOS/uname values to edgelet release OS names.
func NormalizeEdgeletOS(osName string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(osName)) {
	case "linux":
		return "linux", nil
	case "darwin", "macos", "osx":
		return "darwin", nil
	case "windows", "win32":
		return "windows", nil
	default:
		return "", fmt.Errorf("unsupported edgelet OS %q", osName)
	}
}

// NormalizeEdgeletArch maps SDK arch strings to edgelet release arch suffixes.
func NormalizeEdgeletArch(archName string) (string, error) {
	archName = strings.ToLower(strings.TrimSpace(archName))
	if archName == "auto" {
		return "auto", nil
	}
	if _, ok := supportedEdgeletArches[archName]; !ok {
		return "", fmt.Errorf("unsupported edgelet arch %q", archName)
	}
	return archName, nil
}

// EdgeletBinaryArtifact returns the release artifact filename for os/arch.
func EdgeletBinaryArtifact(osName, archName string) (string, error) {
	osName, err := NormalizeEdgeletOS(osName)
	if err != nil {
		return "", err
	}
	archName, err = NormalizeEdgeletArch(archName)
	if err != nil {
		return "", err
	}
	if archName == "auto" {
		return "", fmt.Errorf("edgelet arch must be resolved before building artifact name")
	}
	if osName == "windows" {
		if archName != "amd64" {
			return "", fmt.Errorf("windows edgelet supports amd64 only, got %q", archName)
		}
		return "edgelet-windows-amd64.exe", nil
	}
	return fmt.Sprintf("edgelet-%s-%s", osName, archName), nil
}

// EdgeletBinaryURL builds the GitHub release download URL for a platform binary.
func EdgeletBinaryURL(osName, archName string) (string, error) {
	artifact, err := EdgeletBinaryArtifact(osName, archName)
	if err != nil {
		return "", err
	}
	base := strings.TrimRight(GetEdgeletReleaseBase(), "/")
	version := GetEdgeletBinaryVersion()
	if base == "" || base == "undefined" {
		return "", fmt.Errorf("edgelet release base is not configured")
	}
	if version == "" || version == "undefined" {
		return "", fmt.Errorf("edgelet binary version is not configured")
	}
	return fmt.Sprintf("%s/%s/%s", base, version, artifact), nil
}

// EdgeletChecksumsURL returns the SHA256SUMS manifest URL for the pinned release.
func EdgeletChecksumsURL() (string, error) {
	base := strings.TrimRight(GetEdgeletReleaseBase(), "/")
	version := GetEdgeletBinaryVersion()
	if base == "" || base == "undefined" {
		return "", fmt.Errorf("edgelet release base is not configured")
	}
	if version == "" || version == "undefined" {
		return "", fmt.Errorf("edgelet binary version is not configured")
	}
	return fmt.Sprintf("%s/%s/SHA256SUMS", base, version), nil
}

// ShouldSkipInstallDeps reports whether the deps layer is a no-op for the runtime config.
func ShouldSkipInstallDeps(containerEngine, deploymentType string) bool {
	engine := strings.ToLower(strings.TrimSpace(containerEngine))
	if engine == "" || engine == "edgelet" {
		return true
	}
	_ = deploymentType
	return false
}

// DownloadEdgeletBinary fetches the release binary for os/arch into destPath.
func DownloadEdgeletBinary(osName, archName, destPath string) error {
	url, err := EdgeletBinaryURL(osName, archName)
	if err != nil {
		return err
	}

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download edgelet binary: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download edgelet binary: HTTP %d from %s", resp.StatusCode, url)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("create download dir: %w", err)
	}

	out, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return fmt.Errorf("create edgelet binary file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("write edgelet binary: %w", err)
	}
	return nil
}
