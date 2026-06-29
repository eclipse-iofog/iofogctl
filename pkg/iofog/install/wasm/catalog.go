package wasm

// Pack describes a WASM shim artifact source.
type Pack struct {
	URL    string
	Path   string
	SHA256 string
}

// handlerCandidates maps WASM runtime handler keys to shim binary basenames in
// edgelet searchForRuntimes candidate order (first match wins).
var handlerCandidates = map[string][]string{
	"spin":             {"containerd-shim-spin-v2", "containerd-shim-spin-v1", "spin"},
	"slight":           {"containerd-shim-slight-v1", "containerd-shim-slight-v2", "slight"},
	"lunatic":          {"containerd-shim-lunatic-v1", "containerd-shim-lunatic-v2", "lunatic"},
	"wws":              {"containerd-shim-wws-v1", "containerd-shim-wws-v2", "wws"},
	"wasmedge":         {"containerd-shim-wasmedge-v1", "containerd-shim-wasmedge-v2", "wasmedge"},
	"wasmer":           {"containerd-shim-wasmer-v1", "containerd-shim-wasmer-v2", "wasmer"},
	"wasmtime":         {"containerd-shim-wasmtime-v1", "containerd-shim-wasmtime-v2", "wasmtime"},
	"edgelet-wasmtime": {"containerd-shim-edgelet-v2", "containerd-shim-edgelet-wasm-v2", "containerd-shim-edgelet", "edgelet-wasm"},
}

// Candidates returns ordered shim basenames for handler.
func Candidates(handler string) []string {
	candidates, ok := handlerCandidates[handler]
	if !ok {
		return nil
	}
	out := make([]string, len(candidates))
	copy(out, candidates)
	return out
}

// CanonicalName returns the primary install basename for handler.
func CanonicalName(handler string) string {
	candidates := Candidates(handler)
	if len(candidates) == 0 {
		return ""
	}
	return candidates[0]
}
