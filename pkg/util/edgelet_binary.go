package util

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const edgeletDownloadTimeout = 10 * time.Minute

var edgeletHTTPClient = &http.Client{Timeout: edgeletDownloadTimeout}

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

// ResolveEdgeletBinaryVersion returns packageVersion when set, otherwise the build-time pin.
func ResolveEdgeletBinaryVersion(packageVersion string) string {
	if strings.TrimSpace(packageVersion) != "" {
		return strings.TrimSpace(packageVersion)
	}
	return GetEdgeletBinaryVersion()
}

// EdgeletBinaryURL builds the GitHub release download URL for a platform binary.
// When version is empty, ResolveEdgeletBinaryVersion("") supplies the build-time pin.
func EdgeletBinaryURL(osName, archName, version string) (string, error) {
	artifact, err := EdgeletBinaryArtifact(osName, archName)
	if err != nil {
		return "", err
	}
	base := strings.TrimRight(GetEdgeletReleaseBase(), "/")
	version = ResolveEdgeletBinaryVersion(version)
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

func validateEdgeletDownloadURL(downloadURL, version string) error {
	download, err := url.Parse(downloadURL)
	if err != nil {
		return fmt.Errorf("parse download URL: %w", err)
	}
	if download.Host == "" {
		return fmt.Errorf("download URL missing host")
	}

	base, err := url.Parse(strings.TrimRight(GetEdgeletReleaseBase(), "/"))
	if err != nil {
		return fmt.Errorf("parse edgelet release base: %w", err)
	}
	if base.Host == "" {
		return fmt.Errorf("edgelet release base missing host")
	}
	if download.Host != base.Host {
		return fmt.Errorf("unexpected download host %q", download.Host)
	}
	if base.Scheme == "https" && download.Scheme != "https" {
		return fmt.Errorf("download URL must use HTTPS")
	}
	if download.Scheme != "http" && download.Scheme != "https" {
		return fmt.Errorf("unsupported download scheme %q", download.Scheme)
	}

	version = ResolveEdgeletBinaryVersion(version)
	wantPrefix := strings.TrimRight(base.Path, "/") + "/" + version + "/"
	if !strings.HasPrefix(download.Path, wantPrefix) {
		return fmt.Errorf("unexpected download path %q", download.Path)
	}
	return nil
}

// DownloadEdgeletBinary fetches the release binary for os/arch into destPath.
// When version is empty, ResolveEdgeletBinaryVersion("") supplies the build-time pin.
func DownloadEdgeletBinary(osName, archName, destPath, version string) error {
	return downloadEdgeletBinary(context.Background(), osName, archName, destPath, version, edgeletHTTPClient)
}

func downloadEdgeletBinary(ctx context.Context, osName, archName, destPath, version string, client *http.Client) error {
	downloadURL, err := EdgeletBinaryURL(osName, archName, version)
	if err != nil {
		return err
	}
	if err := validateEdgeletDownloadURL(downloadURL, version); err != nil {
		return fmt.Errorf("validate edgelet download URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return fmt.Errorf("create edgelet download request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download edgelet binary: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download edgelet binary: HTTP %d from %s", resp.StatusCode, downloadURL)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), DirPerm); err != nil {
		return fmt.Errorf("create download dir: %w", err)
	}

	out, err := CreateUserFile(destPath, ExecPerm) // #nosec G302 -- executable edgelet binary
	if err != nil {
		return fmt.Errorf("create edgelet binary file: %w", err)
	}
	defer IgnoreClose(out)

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("write edgelet binary: %w", err)
	}
	return nil
}
