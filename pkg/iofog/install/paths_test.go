package install

import "testing"

func TestWasmLocalStageDirIsolatedPerAgent(t *testing.T) {
	a := WasmLocalStageDir("controller-a")
	b := WasmLocalStageDir("controller-b")
	if a == b {
		t.Fatalf("expected distinct staging dirs, both %q", a)
	}
	if a != "/tmp/edgelet-scripts/wasm/controller-a" {
		t.Fatalf("WasmLocalStageDir(controller-a) = %q", a)
	}
}

func TestSanitizeWasmStageKey(t *testing.T) {
	if got := sanitizeWasmStageKey(""); got != "local" {
		t.Fatalf("empty key = %q, want local", got)
	}
	if got := sanitizeWasmStageKey("CP/Node 1"); got != "cp-node-1" {
		t.Fatalf("sanitized key = %q", got)
	}
}
