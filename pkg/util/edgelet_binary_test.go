package util

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestEdgeletBinaryArtifact(t *testing.T) {
	tests := []struct {
		os      string
		arch    string
		want    string
		wantErr bool
	}{
		{"linux", "amd64", "edgelet-linux-amd64", false},
		{"linux", "arm64", "edgelet-linux-arm64", false},
		{"linux", "arm", "edgelet-linux-arm", false},
		{"linux", "riscv64", "edgelet-linux-riscv64", false},
		{"darwin", "amd64", "edgelet-darwin-amd64", false},
		{"darwin", "arm64", "edgelet-darwin-arm64", false},
		{"windows", "amd64", "edgelet-windows-amd64.exe", false},
		{"windows", "arm64", "", true},
		{"freebsd", "amd64", "", true},
		{"linux", "auto", "", true},
	}

	for _, tt := range tests {
		got, err := EdgeletBinaryArtifact(tt.os, tt.arch)
		if tt.wantErr {
			if err == nil {
				t.Fatalf("EdgeletBinaryArtifact(%q, %q) expected error", tt.os, tt.arch)
			}
			continue
		}
		if err != nil {
			t.Fatalf("EdgeletBinaryArtifact(%q, %q): %v", tt.os, tt.arch, err)
		}
		if got != tt.want {
			t.Fatalf("EdgeletBinaryArtifact(%q, %q) = %q, want %q", tt.os, tt.arch, got, tt.want)
		}
	}
}

func TestEdgeletBinaryURL(t *testing.T) {
	SetEdgeletReleaseBaseForTest("https://github.com/Datasance/edgelet/releases/download")
	SetEdgeletBinaryVersionForTest("v1.0.0-rc.8")
	t.Cleanup(ResetEdgeletReleaseBaseForTest)
	t.Cleanup(ResetEdgeletBinaryVersionForTest)

	got, err := EdgeletBinaryURL("linux", "amd64")
	if err != nil {
		t.Fatalf("EdgeletBinaryURL: %v", err)
	}
	want := "https://github.com/Datasance/edgelet/releases/download/v1.0.0-rc.8/edgelet-linux-amd64"
	if got != want {
		t.Fatalf("EdgeletBinaryURL = %q, want %q", got, want)
	}
	if err := validateEdgeletDownloadURL(got); err != nil {
		t.Fatalf("validateEdgeletDownloadURL(%q): %v", got, err)
	}
}

func TestShouldSkipInstallDeps(t *testing.T) {
	tests := []struct {
		engine   string
		deploy   string
		expected bool
	}{
		{"edgelet", "native", true},
		{"edgelet", "container", true},
		{"", "native", true},
		{"docker", "native", false},
		{"podman", "container", false},
	}

	for _, tt := range tests {
		if got := ShouldSkipInstallDeps(tt.engine, tt.deploy); got != tt.expected {
			t.Fatalf("ShouldSkipInstallDeps(%q, %q) = %v, want %v", tt.engine, tt.deploy, got, tt.expected)
		}
	}
}

func TestDownloadEdgeletBinaryGitHubReleasePath(t *testing.T) {
	const releasePath = "/Datasance/edgelet/releases/download/v1.0.0-rc.8/edgelet-linux-arm64"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != releasePath {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte("edgelet-binary-bytes"))
	}))
	defer server.Close()

	SetEdgeletReleaseBaseForTest(server.URL + "/Datasance/edgelet/releases/download")
	SetEdgeletBinaryVersionForTest("v1.0.0-rc.8")
	t.Cleanup(ResetEdgeletReleaseBaseForTest)
	t.Cleanup(ResetEdgeletBinaryVersionForTest)

	dir := t.TempDir()
	dest := filepath.Join(dir, "edgelet-linux-arm64")

	if err := DownloadEdgeletBinary("linux", "arm64", dest); err != nil {
		t.Fatalf("DownloadEdgeletBinary: %v", err)
	}
}

func TestDownloadEdgeletBinary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1.0.0-rc.8/edgelet-linux-amd64" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte("edgelet-binary-bytes"))
	}))
	defer server.Close()

	SetEdgeletReleaseBaseForTest(server.URL)
	SetEdgeletBinaryVersionForTest("v1.0.0-rc.8")
	t.Cleanup(ResetEdgeletReleaseBaseForTest)
	t.Cleanup(ResetEdgeletBinaryVersionForTest)

	dir := t.TempDir()
	dest := filepath.Join(dir, "edgelet-linux-amd64")

	if err := DownloadEdgeletBinary("linux", "amd64", dest); err != nil {
		t.Fatalf("DownloadEdgeletBinary: %v", err)
	}

	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read downloaded binary: %v", err)
	}
	if string(data) != "edgelet-binary-bytes" {
		t.Fatalf("unexpected binary contents: %q", string(data))
	}
}
