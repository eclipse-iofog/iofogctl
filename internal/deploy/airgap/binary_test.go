package deployairgap

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func TestEnsureEdgeletBinaryUsesCache(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1.0.0-rc.6/edgelet-linux-amd64" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte("cached-edgelet-binary"))
	}))
	defer server.Close()

	util.SetEdgeletReleaseBaseForTest(server.URL)
	util.SetEdgeletBinaryVersionForTest("v1.0.0-rc.6")
	t.Cleanup(func() {
		util.ResetEdgeletReleaseBaseForTest()
		util.ResetEdgeletBinaryVersionForTest()
	})

	config.Init(t.TempDir())
	namespace := "default"

	first, err := EnsureEdgeletBinary(context.Background(), namespace, "linux", "amd64")
	if err != nil {
		t.Fatalf("EnsureEdgeletBinary first: %v", err)
	}
	data, err := os.ReadFile(first)
	if err != nil {
		t.Fatalf("read cached binary: %v", err)
	}
	if string(data) != "cached-edgelet-binary" {
		t.Fatalf("unexpected binary contents: %q", string(data))
	}

	second, err := EnsureEdgeletBinary(context.Background(), namespace, "linux", "amd64")
	if err != nil {
		t.Fatalf("EnsureEdgeletBinary second: %v", err)
	}
	if second != first {
		t.Fatalf("expected cache reuse at %q, got %q", first, second)
	}

	wantPath := config.GetAirgapBinaryCachePath(namespace, "linux", "amd64")
	if first != wantPath {
		t.Fatalf("cache path = %q, want %q", first, wantPath)
	}

	metaPath := filepath.Join(filepath.Dir(first), binaryMetadataFilename)
	if _, err := os.Stat(metaPath); err != nil {
		t.Fatalf("metadata file missing: %v", err)
	}
}

func TestGetAirgapBinaryCachePathWindows(t *testing.T) {
	config.Init(t.TempDir())
	got := config.GetAirgapBinaryCachePath("default", "windows", "amd64")
	if !strings.Contains(got, "edgelet-windows-amd64.exe") {
		t.Fatalf("cache path = %q, want windows executable artifact name", got)
	}
}
