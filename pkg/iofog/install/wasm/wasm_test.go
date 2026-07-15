package wasm

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func initWasmTestCache(t *testing.T) {
	t.Helper()
	SetCacheRootForTest(t.TempDir())
	t.Cleanup(ResetCacheRootForTest)
}

func TestExtractTarFindsSpinShim(t *testing.T) {
	t.Parallel()

	flat := writeTarGz(t, map[string][]byte{
		"containerd-shim-spin-v2": []byte("spin-shim-binary"),
	})
	path, name, err := extractArtifact(flat, "spin")
	require.NoError(t, err)
	require.Equal(t, "containerd-shim-spin-v2", name)
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, []byte("spin-shim-binary"), data)
	t.Cleanup(func() { _ = os.Remove(path) })

	nested := writeTarGz(t, map[string][]byte{
		"nested/containerd-shim-spin-v2": []byte("nested-spin-shim"),
	})
	path, name, err = extractArtifact(nested, "spin")
	require.NoError(t, err)
	require.Equal(t, "containerd-shim-spin-v2", name)
	data, err = os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, []byte("nested-spin-shim"), data)
	t.Cleanup(func() { _ = os.Remove(path) })
}

func TestExtractRawELF(t *testing.T) {
	t.Parallel()

	rawPath := filepath.Join(t.TempDir(), "raw-shim")
	require.NoError(t, os.WriteFile(rawPath, minimalELF([]byte("raw-shim-payload")), 0o755))

	path, name, err := extractArtifact(rawPath, "spin")
	require.NoError(t, err)
	require.Equal(t, "containerd-shim-spin-v2", name)
	require.Equal(t, rawPath, path)
}

func TestResolveSkipsEmptyMap(t *testing.T) {
	t.Parallel()

	initWasmTestCache(t)
	got, err := ResolveWasmArtifacts(context.Background(), "default", "linux/amd64", nil, false)
	require.NoError(t, err)
	require.Nil(t, got)

	got, err = ResolveWasmArtifacts(context.Background(), "default", "linux/amd64", map[string]Pack{}, false)
	require.NoError(t, err)
	require.Nil(t, got)
}

func TestCatalogCandidateOrder(t *testing.T) {
	t.Parallel()

	require.Equal(t, []string{
		"containerd-shim-spin-v2",
		"containerd-shim-spin-v1",
		"spin",
	}, Candidates("spin"))
	require.Equal(t, []string{
		"containerd-shim-edgelet-v2",
		"containerd-shim-edgelet-wasm-v2",
		"containerd-shim-edgelet",
		"edgelet-wasm",
	}, Candidates("edgelet-wasmtime"))

	archive := writeTarGz(t, map[string][]byte{
		"containerd-shim-spin-v1": []byte("v1"),
		"containerd-shim-spin-v2": []byte("v2"),
	})
	_, name, err := extractArtifact(archive, "spin")
	require.NoError(t, err)
	require.Equal(t, "containerd-shim-spin-v2", name)

	v1Only := writeTarGz(t, map[string][]byte{
		"containerd-shim-spin-v1": []byte("v1-only"),
	})
	_, name, err = extractArtifact(v1Only, "spin")
	require.NoError(t, err)
	require.Equal(t, "containerd-shim-spin-v1", name)
}

func TestWasmCatalogMatchesEdgelet(t *testing.T) {
	t.Parallel()

	edgeletCandidates := map[string][]string{
		"spin":             {"containerd-shim-spin-v2", "containerd-shim-spin-v1", "spin"},
		"slight":           {"containerd-shim-slight-v1", "containerd-shim-slight-v2", "slight"},
		"lunatic":          {"containerd-shim-lunatic-v1", "containerd-shim-lunatic-v2", "lunatic"},
		"wws":              {"containerd-shim-wws-v1", "containerd-shim-wws-v2", "wws"},
		"wasmedge":         {"containerd-shim-wasmedge-v1", "containerd-shim-wasmedge-v2", "wasmedge"},
		"wasmer":           {"containerd-shim-wasmer-v1", "containerd-shim-wasmer-v2", "wasmer"},
		"wasmtime":         {"containerd-shim-wasmtime-v1", "containerd-shim-wasmtime-v2", "wasmtime"},
		"edgelet-wasmtime": {"containerd-shim-edgelet-v2", "containerd-shim-edgelet-wasm-v2", "containerd-shim-edgelet", "edgelet-wasm"},
	}

	for handler, want := range edgeletCandidates {
		require.Equal(t, want, Candidates(handler), "handler %q", handler)
	}
}

func TestResolveWasmArtifactsFromPath(t *testing.T) {
	initWasmTestCache(t)

	archive := writeTarGz(t, map[string][]byte{
		"containerd-shim-spin-v2": []byte("resolved-spin-shim"),
	})

	staged, err := ResolveWasmArtifacts(context.Background(), "default", "linux/amd64", map[string]Pack{
		"spin": {Path: archive},
	}, false)
	require.NoError(t, err)
	require.Len(t, staged, 1)
	require.Equal(t, "spin", staged[0].Handler)
	require.Equal(t, "containerd-shim-spin-v2", staged[0].CanonicalName)
	require.FileExists(t, staged[0].LocalPath)

	data, err := os.ReadFile(staged[0].LocalPath)
	require.NoError(t, err)
	require.Equal(t, []byte("resolved-spin-shim"), data)
	require.NotEmpty(t, staged[0].SHA256)
}

func TestResolveWasmArtifactsSHA256Mismatch(t *testing.T) {
	initWasmTestCache(t)

	archive := writeTarGz(t, map[string][]byte{
		"containerd-shim-spin-v2": []byte("resolved-spin-shim"),
	})

	_, err := ResolveWasmArtifacts(context.Background(), "default", "linux/amd64", map[string]Pack{
		"spin": {
			Path:   archive,
			SHA256: "0000000000000000000000000000000000000000000000000000000000000000",
		},
	}, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "sha256 mismatch")
}

func TestResolveWasmArtifactsFromURL(t *testing.T) {
	payload := writeTarGzBytes(t, map[string][]byte{
		"containerd-shim-spin-v2": []byte("downloaded-spin-shim"),
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(payload)))
		_, _ = w.Write(payload)
	}))
	defer server.Close()

	oldClient := httpClient
	httpClient = server.Client()
	t.Cleanup(func() { httpClient = oldClient })

	initWasmTestCache(t)

	staged, err := ResolveWasmArtifacts(context.Background(), "default", "linux/amd64", map[string]Pack{
		"spin": {URL: server.URL + "/spin.tar.gz"},
	}, false)
	require.NoError(t, err)
	require.Len(t, staged, 1)
	require.Equal(t, "containerd-shim-spin-v2", staged[0].CanonicalName)

	second, err := ResolveWasmArtifacts(context.Background(), "default", "linux/amd64", map[string]Pack{
		"spin": {URL: server.URL + "/spin.tar.gz"},
	}, false)
	require.NoError(t, err)
	require.Equal(t, staged[0].LocalPath, second[0].LocalPath)
}

func TestResolveWasmArtifactsValidSHA256(t *testing.T) {
	initWasmTestCache(t)

	archive := writeTarGz(t, map[string][]byte{
		"containerd-shim-spin-v2": []byte("resolved-spin-shim"),
	})
	sum := sha256.Sum256([]byte("resolved-spin-shim"))

	staged, err := ResolveWasmArtifacts(context.Background(), "default", "linux/amd64", map[string]Pack{
		"spin": {
			Path:   archive,
			SHA256: hex.EncodeToString(sum[:]),
		},
	}, false)
	require.NoError(t, err)
	require.Len(t, staged, 1)
}

func TestResolveWasmArtifactsCacheReuse(t *testing.T) {
	initWasmTestCache(t)

	archive := writeTarGz(t, map[string][]byte{
		"containerd-shim-spin-v2": []byte("cached-content"),
	})
	pack := map[string]Pack{"spin": {Path: archive}}

	first, err := ResolveWasmArtifacts(context.Background(), "default", "linux/amd64", pack, false)
	require.NoError(t, err)

	second, err := ResolveWasmArtifacts(context.Background(), "default", "linux/amd64", pack, false)
	require.NoError(t, err)
	require.Equal(t, first[0].LocalPath, second[0].LocalPath)
	require.Equal(t, first[0].SHA256, second[0].SHA256)
}

func TestResolveWasmArtifactsChangedWhenInstalledDiffers(t *testing.T) {
	initWasmTestCache(t)
	installDir := t.TempDir()
	oldInstallDir := localInstallDir
	localInstallDir = installDir
	t.Cleanup(func() { localInstallDir = oldInstallDir })

	archive := writeTarGz(t, map[string][]byte{
		"containerd-shim-spin-v2": []byte("new-content"),
	})
	destPath := filepath.Join(installDir, "containerd-shim-spin-v2")
	require.NoError(t, os.WriteFile(destPath, []byte("old-content"), 0o755))

	staged, err := ResolveWasmArtifacts(context.Background(), "default", "linux/amd64", map[string]Pack{
		"spin": {Path: archive},
	}, false)
	require.NoError(t, err)
	require.True(t, staged[0].Changed)
}

func TestResolveWasmArtifactsDownloadError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	oldClient := httpClient
	httpClient = server.Client()
	t.Cleanup(func() { httpClient = oldClient })

	initWasmTestCache(t)

	_, err := ResolveWasmArtifacts(context.Background(), "default", "linux/amd64", map[string]Pack{
		"spin": {URL: server.URL + "/missing.tar.gz"},
	}, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "HTTP 404")
}

func TestCopyExtractedBinary(t *testing.T) {
	src := filepath.Join(t.TempDir(), "src")
	dest := filepath.Join(t.TempDir(), "bin", "containerd-shim-spin-v2")
	require.NoError(t, os.WriteFile(src, []byte("copy-me"), 0o644))
	require.NoError(t, copyExtractedBinary(src, dest))
	require.Equal(t, []byte("copy-me"), readFile(t, dest))

	info, err := os.Stat(dest)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o755), info.Mode().Perm())
}

func TestInstallLocalBinaries(t *testing.T) {
	destDir := t.TempDir()
	srcDir := t.TempDir()
	srcPath := filepath.Join(srcDir, "containerd-shim-spin-v2")
	require.NoError(t, os.WriteFile(srcPath, []byte("install-me"), 0o755))

	installed, err := InstallLocalBinaries([]StagedBinary{{
		CanonicalName: "containerd-shim-spin-v2",
		LocalPath:     srcPath,
	}}, destDir)
	require.NoError(t, err)
	require.Len(t, installed, 1)

	info, err := os.Stat(installed[0])
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o755), info.Mode().Perm())
}

func TestResolveWasmArtifactsRawELF(t *testing.T) {
	initWasmTestCache(t)
	installDir := t.TempDir()
	oldInstallDir := localInstallDir
	localInstallDir = installDir
	t.Cleanup(func() { localInstallDir = oldInstallDir })

	rawPath := filepath.Join(t.TempDir(), "raw-shim")
	require.NoError(t, os.WriteFile(rawPath, minimalELF([]byte("raw-resolve")), 0o755))

	staged, err := ResolveWasmArtifacts(context.Background(), "default", "linux/amd64", map[string]Pack{
		"spin": {Path: rawPath},
	}, false)
	require.NoError(t, err)
	require.Len(t, staged, 1)
	require.True(t, staged[0].Changed)
	require.Equal(t, "containerd-shim-spin-v2", staged[0].CanonicalName)

	destPath := filepath.Join(installDir, staged[0].CanonicalName)
	require.NoError(t, os.WriteFile(destPath, minimalELF([]byte("raw-resolve")), 0o755))

	staged, err = ResolveWasmArtifacts(context.Background(), "default", "linux/amd64", map[string]Pack{
		"spin": {Path: rawPath},
	}, false)
	require.NoError(t, err)
	require.False(t, staged[0].Changed)
}

func TestResolveWasmArtifactsAirgapDownloadsFromURL(t *testing.T) {
	payload := writeTarGzBytes(t, map[string][]byte{
		"containerd-shim-spin-v2": []byte("airgap-downloaded-spin-shim"),
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(payload)))
		_, _ = w.Write(payload)
	}))
	defer server.Close()

	oldClient := httpClient
	httpClient = server.Client()
	t.Cleanup(func() { httpClient = oldClient })

	initWasmTestCache(t)

	staged, err := ResolveWasmArtifacts(context.Background(), "default", "linux/amd64", map[string]Pack{
		"spin": {URL: server.URL + "/spin.tar.gz"},
	}, true)
	require.NoError(t, err)
	require.Len(t, staged, 1)
	require.Equal(t, "containerd-shim-spin-v2", staged[0].CanonicalName)
	require.Equal(t, []byte("airgap-downloaded-spin-shim"), readFile(t, staged[0].LocalPath))
}

func TestResolveWasmArtifactsAirgapUsesCache(t *testing.T) {
	initWasmTestCache(t)

	archive := writeTarGz(t, map[string][]byte{
		"containerd-shim-spin-v2": []byte("airgap-spin-shim"),
	})
	sourceURL := "https://example.invalid/spin.tar.gz"
	cacheSource := sourceCachePath("default", "spin", "linux", "amd64")
	require.NoError(t, os.MkdirAll(filepath.Dir(cacheSource), 0o700))
	require.NoError(t, copyFile(archive, cacheSource))
	checksum, _, err := fileSHA256(cacheSource)
	require.NoError(t, err)
	require.NoError(t, saveCacheMetadata(metadataPath("default", "spin", "linux", "amd64"), cacheMetadata{
		Handler:        "spin",
		OS:             "linux",
		Arch:           "amd64",
		SourceURL:      sourceURL,
		SourceChecksum: checksum,
		CanonicalName:  "containerd-shim-spin-v2",
	}))

	staged, err := ResolveWasmArtifacts(context.Background(), "default", "linux/amd64", map[string]Pack{
		"spin": {URL: sourceURL},
	}, true)
	require.NoError(t, err)
	require.Len(t, staged, 1)
	require.Equal(t, []byte("airgap-spin-shim"), readFile(t, staged[0].LocalPath))
}

func TestResolveWasmArtifactsInvalidPlatform(t *testing.T) {
	initWasmTestCache(t)

	_, err := ResolveWasmArtifacts(context.Background(), "default", "invalid", map[string]Pack{
		"spin": {Path: writeTarGz(t, map[string][]byte{"containerd-shim-spin-v2": []byte("x")})},
	}, false)
	require.Error(t, err)
}

func TestFormatDetection(t *testing.T) {
	t.Parallel()

	require.False(t, isGzipTar(filepath.Join(t.TempDir(), "missing")))
	require.False(t, isRawELF(filepath.Join(t.TempDir(), "missing")))

	plain := filepath.Join(t.TempDir(), "plain.txt")
	require.NoError(t, os.WriteFile(plain, []byte("hello"), 0o644))
	require.False(t, isGzipTar(plain))
	require.False(t, isRawELF(plain))
	require.True(t, isGzipTar(writeTarGz(t, map[string][]byte{"x": []byte("y")})))

	elfPath := filepath.Join(t.TempDir(), "elf")
	require.NoError(t, os.WriteFile(elfPath, minimalELF(nil), 0o755))
	require.True(t, isRawELF(elfPath))
}

func TestResolveRefreshesStaleCache(t *testing.T) {
	initWasmTestCache(t)

	archive := writeTarGz(t, map[string][]byte{
		"containerd-shim-spin-v2": []byte("fresh-content"),
	})
	pack := map[string]Pack{"spin": {Path: archive}}

	staged, err := ResolveWasmArtifacts(context.Background(), "default", "linux/amd64", pack, false)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(staged[0].LocalPath, []byte("stale"), 0o755))

	staged, err = ResolveWasmArtifacts(context.Background(), "default", "linux/amd64", pack, false)
	require.NoError(t, err)
	require.Equal(t, []byte("fresh-content"), readFile(t, staged[0].LocalPath))
}

func TestExtractArtifactErrors(t *testing.T) {
	t.Parallel()

	_, _, err := extractArtifact(writeTarGz(t, map[string][]byte{"other-binary": []byte("x")}), "spin")
	require.Error(t, err)

	unsupported := filepath.Join(t.TempDir(), "plain.txt")
	require.NoError(t, os.WriteFile(unsupported, []byte("not-a-shim"), 0o644))
	_, _, err = extractArtifact(unsupported, "spin")
	require.Error(t, err)

	_, _, err = extractArtifact(unsupported, "unknown")
	require.Error(t, err)

	emptyTar := writeTarGz(t, map[string][]byte{})
	_, _, err = extractArtifact(emptyTar, "spin")
	require.Error(t, err)

	badGzip := filepath.Join(t.TempDir(), "bad.tar.gz")
	require.NoError(t, os.WriteFile(badGzip, []byte{0x1f, 0x8b, 0x00, 0x00}, 0o644))
	_, _, err = extractArtifact(badGzip, "spin")
	require.Error(t, err)
}

func TestInstallLocalBinariesDefaultDest(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("requires root to write /usr/local/bin")
	}
	srcPath := filepath.Join(t.TempDir(), "containerd-shim-spin-v2")
	require.NoError(t, os.WriteFile(srcPath, []byte("root-install"), 0o755))
	t.Cleanup(func() { _ = os.Remove(filepath.Join(defaultInstallDir, "containerd-shim-spin-v2")) })

	installed, err := InstallLocalBinaries([]StagedBinary{{
		CanonicalName: "containerd-shim-spin-v2",
		LocalPath:     srcPath,
	}}, "")
	require.NoError(t, err)
	require.Len(t, installed, 1)
}

func TestWasmCacheDir(t *testing.T) {
	initWasmTestCache(t)
	got := CacheDir("default", "spin", "linux", "amd64")
	require.Contains(t, got, filepath.Join("airgap-binaries", "default", "wasm", "spin", "linux-amd64"))
}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return data
}

func writeTarGz(t *testing.T, entries map[string][]byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fixture.tar.gz")
	require.NoError(t, os.WriteFile(path, writeTarGzBytes(t, entries), 0o644))
	return path
}

func writeTarGzBytes(t *testing.T, entries map[string][]byte) []byte {
	t.Helper()

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, content := range entries {
		require.NoError(t, tw.WriteHeader(&tar.Header{
			Name: name,
			Mode: 0o755,
			Size: int64(len(content)),
		}))
		_, err := tw.Write(content)
		require.NoError(t, err)
	}
	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())
	return buf.Bytes()
}

func minimalELF(payload []byte) []byte {
	header := []byte{
		0x7f, 'E', 'L', 'F',
		2, // 64-bit
		1, // little endian
		1, // current version
		0, // sysv
		0, 0, 0, 0, 0, 0, 0, 0,
		2, 0, // executable
		0x3e, 0,
	}
	out := make([]byte, 0, len(header)+len(payload))
	out = append(out, header...)
	out = append(out, payload...)
	return out
}
