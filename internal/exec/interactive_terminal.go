package exec

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"golang.org/x/term"
)

const (
	execStdinChunk       = 1024
	minExecTermCols      = 2
	minExecTermRows      = 1
	maxExecTermDimension = 4096
)

// execSession is the controller exec socket. *client.ExecSession implements it.
type execSession interface {
	Read() (*client.ExecFrame, error)
	WriteStdin([]byte) error
	WriteControl([]byte) error
	Close() error
}

type interactiveTerminal struct {
	session             execSession
	ctx                 context.Context
	cancel              context.CancelFunc
	cleanupOnce         sync.Once
	stdoutMutex         sync.Mutex
	writeMu             sync.Mutex
	userInitiatedClose  bool
	userInitiatedCloseM sync.Mutex
}

func newInteractiveTerminal(session execSession) *interactiveTerminal {
	ctx, cancel := context.WithCancel(context.Background())
	return &interactiveTerminal{
		session: session,
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (t *interactiveTerminal) writeToStdout(data []byte) {
	t.stdoutMutex.Lock()
	defer t.stdoutMutex.Unlock()
	_, _ = os.Stdout.Write(data)
}

func (t *interactiveTerminal) cleanup() {
	t.cleanupOnce.Do(func() {
		if t.session != nil {
			t.writeMu.Lock()
			_ = t.session.Close()
			t.writeMu.Unlock()
		}
	})
}

func (t *interactiveTerminal) setUserInitiatedClose() {
	t.userInitiatedCloseM.Lock()
	t.userInitiatedClose = true
	t.userInitiatedCloseM.Unlock()
}

func (t *interactiveTerminal) isUserInitiatedClose() bool {
	t.userInitiatedCloseM.Lock()
	defer t.userInitiatedCloseM.Unlock()
	return t.userInitiatedClose
}

// handleInput forwards one stdin chunk unchanged. The remote PTY echoes and
// interprets keys, including Ctrl+C. A write error ends the local relay.
func (t *interactiveTerminal) handleInput(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	t.writeMu.Lock()
	err := t.session.WriteStdin(data)
	t.writeMu.Unlock()
	if err != nil {
		t.cancel()
		return true
	}
	return false
}

func (t *interactiveTerminal) sendResize() {
	cols, rows, err := terminalSize()
	if err != nil {
		return
	}
	_ = t.writeResize(cols, rows)
}

func (t *interactiveTerminal) writeResize(cols, rows uint32) error {
	payload, err := resizeControlPayload(cols, rows)
	if err != nil {
		return err
	}
	t.writeMu.Lock()
	defer t.writeMu.Unlock()
	return t.session.WriteControl(payload)
}

func (t *interactiveTerminal) copyStdin() {
	buf := make([]byte, execStdinChunk)
	for {
		if t.ctx.Err() != nil {
			return
		}
		n, readErr := os.Stdin.Read(buf)
		if n > 0 {
			chunk := append([]byte(nil), buf[:n]...)
			if t.handleInput(chunk) {
				return
			}
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				t.setUserInitiatedClose()
			}
			t.cancel()
			return
		}
	}
}

func (t *interactiveTerminal) shouldPrintFrame(frame *client.ExecFrame) bool {
	switch frame.Type {
	case client.ExecMessageStdout, client.ExecMessageStderr:
		return true
	case client.ExecMessageActivation, client.ExecMessageControl, client.ExecMessageClose:
		return false
	default:
		return false
	}
}

func (t *interactiveTerminal) run() error {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return err
	}
	defer func() {
		if oldState != nil {
			_ = term.Restore(fd, oldState)
		}
	}()

	t.sendResize()
	stopResize := watchTerminalResize(t.sendResize)
	defer stopResize()

	return t.runPlatform(func() {
		go t.copyStdin()
	})
}

func (t *interactiveTerminal) runPlatform(startInput func()) error {
	defer t.cleanup()

	errCh := make(chan error, 1)
	go func() {
		for {
			select {
			case <-t.ctx.Done():
				return
			default:
				frame, err := t.session.Read()
				if err != nil {
					if isNormalExecClose(err, t.isUserInitiatedClose()) {
						t.cancel()
						return
					}
					errCh <- err
					t.cancel()
					return
				}
				if frame == nil {
					t.cancel()
					return
				}
				if frame.Type == client.ExecMessageClose {
					t.cancel()
					return
				}
				if t.shouldPrintFrame(frame) && len(frame.Data) > 0 {
					t.writeToStdout(frame.Data)
				}
			}
		}
	}()

	startInput()

	select {
	case <-t.ctx.Done():
		return nil
	case err := <-errCh:
		if isNormalExecClose(err, t.isUserInitiatedClose()) {
			return nil
		}
		return err
	}
}

func terminalSize() (uint32, uint32, error) {
	fd := int(os.Stdout.Fd())
	if !term.IsTerminal(fd) {
		fd = int(os.Stdin.Fd())
	}
	if !term.IsTerminal(fd) {
		return 0, 0, errors.New("stdout is not a terminal")
	}
	cols, rows, err := term.GetSize(fd)
	if err != nil {
		return 0, 0, err
	}
	if cols <= 0 || rows <= 0 {
		return 0, 0, errors.New("terminal size is unset")
	}
	return uint32(cols), uint32(rows), nil // #nosec G115 -- terminal dimensions from term.GetSize are small positive ints
}

func resizeControlPayload(cols, rows uint32) ([]byte, error) {
	if cols < minExecTermCols || rows < minExecTermRows || cols > maxExecTermDimension || rows > maxExecTermDimension {
		return nil, fmt.Errorf("invalid terminal size %dx%d", cols, rows)
	}
	return json.Marshal(struct {
		Type string `json:"type"`
		Cols uint32 `json:"cols"`
		Rows uint32 `json:"rows"`
	}{
		Type: "resize",
		Cols: cols,
		Rows: rows,
	})
}
