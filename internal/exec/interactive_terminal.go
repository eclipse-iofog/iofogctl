package exec

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type interactiveTerminal struct {
	session             *client.ExecSession
	ctx                 context.Context
	cancel              context.CancelFunc
	cleanupOnce         sync.Once
	stdoutMutex         sync.Mutex
	lastCtrlCTime       time.Time
	lineBuffer          string
	userInitiatedClose  bool
	userInitiatedCloseM sync.Mutex
}

func newInteractiveTerminal(session *client.ExecSession) *interactiveTerminal {
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
	util.WriteStdout(data)
}

func (t *interactiveTerminal) cleanup() {
	t.cleanupOnce.Do(func() {
		if t.session != nil {
			_ = t.session.Close()
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

func (t *interactiveTerminal) sendExecClose() {
	t.setUserInitiatedClose()
	t.cancel()
}

func (t *interactiveTerminal) handleInput(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	switch rune(data[0]) {
	case 0x03: // Ctrl+C
		if time.Since(t.lastCtrlCTime) < time.Second {
			t.setUserInitiatedClose()
			t.cancel()
			t.cleanup()
			t.writeToStdout([]byte("\nExiting...\n"))
			return true
		}
		t.lastCtrlCTime = time.Now()
		t.writeToStdout([]byte("^C\n"))
	}

	if err := t.session.WriteStdin(data); err != nil {
		t.cancel()
		return true
	}

	input := string(data)
	t.lineBuffer += input
	if isShellExitInput(input, t.lineBuffer) {
		t.lineBuffer = ""
		t.sendExecClose()
		t.cancel()
		return true
	}
	if strings.ContainsAny(input, "\r\n") {
		t.lineBuffer = ""
	}

	return false
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
