//go:build windows
// +build windows

package exec

import (
	"os"

	"golang.org/x/term"
)

func (t *interactiveTerminal) run() error {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	defer func() {
		if oldState != nil {
			_ = term.Restore(int(os.Stdin.Fd()), oldState)
		}
	}()

	return t.runPlatform(func() {
		go func() {
			buf := make([]byte, 1)
			for {
				select {
				case <-t.ctx.Done():
					return
				default:
					n, readErr := os.Stdin.Read(buf)
					if readErr != nil || n == 0 {
						continue
					}
					if t.handleInput(buf[:n]) {
						return
					}
				}
			}
		}()
	})
}
