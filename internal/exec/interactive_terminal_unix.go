//go:build !windows
// +build !windows

package exec

import (
	"os"
	"os/signal"
	"syscall"
)

// watchTerminalResize invokes send on SIGWINCH. The returned stop function
// waits until the watcher has exited.
func watchTerminalResize(send func()) func() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for range sigCh {
			send()
		}
	}()
	return func() {
		signal.Stop(sigCh)
		close(sigCh)
		<-done
	}
}
