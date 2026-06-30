package resource

// AllowedWasmHandlers lists edgelet runtime handler keys for package.wasm.
var AllowedWasmHandlers = []string{
	"spin",
	"slight",
	"lunatic",
	"wws",
	"wasmedge",
	"wasmer",
	"wasmtime",
	"edgelet-wasmtime",
}

var allowedWasmHandlerSet = func() map[string]struct{} {
	set := make(map[string]struct{}, len(AllowedWasmHandlers))
	for _, handler := range AllowedWasmHandlers {
		set[handler] = struct{}{}
	}
	return set
}()

func isAllowedWasmHandler(name string) bool {
	_, ok := allowedWasmHandlerSet[name]
	return ok
}
