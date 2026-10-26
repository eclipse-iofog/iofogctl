//go:build windows
// +build windows

package exec

// watchTerminalResize is a no-op on Windows. The initial size is sent by run.
func watchTerminalResize(func()) func() {
	return func() {}
}
