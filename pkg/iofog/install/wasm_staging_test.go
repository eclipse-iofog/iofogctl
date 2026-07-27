package install

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPruneWasmStageDirRemovesOnlyStaleFiles(t *testing.T) {
	dir := t.TempDir()
	keepPath := filepath.Join(dir, "containerd-shim-spin-v2")
	stalePath := filepath.Join(dir, "containerd-shim-wasmtime-v2")
	for _, path := range []string{keepPath, stalePath} {
		if err := os.WriteFile(path, []byte("bin"), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	if err := pruneWasmStageDir(dir, map[string]struct{}{"containerd-shim-spin-v2": {}}); err != nil {
		t.Fatalf("pruneWasmStageDir: %v", err)
	}
	if _, err := os.Stat(keepPath); err != nil {
		t.Fatalf("kept binary removed: %v", err)
	}
	if _, err := os.Stat(stalePath); !os.IsNotExist(err) {
		t.Fatalf("expected stale binary removed, stat err=%v", err)
	}
}

func TestPruneWasmStageDirDoesNotRemoveManifest(t *testing.T) {
	dir := t.TempDir()
	manifest := filepath.Join(dir, "manifest.json")
	if err := os.WriteFile(manifest, []byte("[]"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := pruneWasmStageDir(dir, map[string]struct{}{}); err != nil {
		t.Fatalf("pruneWasmStageDir: %v", err)
	}
	if _, err := os.Stat(manifest); err != nil {
		t.Fatalf("manifest.json removed: %v", err)
	}
}

func TestReadWasmManifest(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")
	if err := os.WriteFile(path, []byte(`[{"handler":"spin","src":"/tmp/spin","name":"containerd-shim-spin-v2"}]`), 0o644); err != nil {
		t.Fatal(err)
	}
	entries, err := readWasmManifest(path)
	if err != nil {
		t.Fatalf("readWasmManifest: %v", err)
	}
	if len(entries) != 1 || entries[0].Handler != "spin" {
		t.Fatalf("unexpected entries: %#v", entries)
	}
}
