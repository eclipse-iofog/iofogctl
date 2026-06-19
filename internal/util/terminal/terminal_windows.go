//go:build windows
// +build windows

package terminal

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	ws "github.com/eclipse-iofog/iofogctl/internal/util/websocket"
	"golang.org/x/term"
)

type Terminal struct {
	wsClient    *ws.Client
	oldState    *term.State
	history     []string
	histIdx     int
	inputBuffer []rune
	cursorPos   int
	resizeCh    chan os.Signal
	// lastCommand   string
	lastCtrlCTime time.Time
	prompt        string
	stdoutMutex   sync.Mutex
	ctx           context.Context
	cancel        context.CancelFunc
	cleanupOnce   sync.Once
	// isEditorMode  bool
}

// var promptPattern = regexp.MustCompile(`(?m)^.*[@].*[$#] ?$`)

func NewTerminal(wsClient *ws.Client) *Terminal {
	ctx, cancel := context.WithCancel(context.Background())
	return &Terminal{
		wsClient:    wsClient,
		history:     make([]string, 0, 1000),
		histIdx:     -1,
		inputBuffer: make([]rune, 0, 1024),
		resizeCh:    make(chan os.Signal, 1),
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (t *Terminal) writeToStdout(data []byte) {
	t.stdoutMutex.Lock()
	defer t.stdoutMutex.Unlock()
	os.Stdout.Write(data)
	os.Stdout.Sync()
}

func (t *Terminal) handleInput(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	// Handle critical control sequences locally
	r := rune(data[0])
	switch r {
	case 0x03: // Ctrl+C
		if time.Since(t.lastCtrlCTime) < time.Second {
			t.cancel()
			t.wsClient.Close()
			t.writeToStdout([]byte("\nExiting...\n"))
			return true
		}
		t.lastCtrlCTime = time.Now()
		t.inputBuffer = t.inputBuffer[:0]
		t.cursorPos = 0
		t.redrawInputLine()
		t.writeToStdout([]byte("^C\n"))
	case 0x04: // Ctrl+D
		if len(t.inputBuffer) == 0 {
			t.cancel()
			t.wsClient.Close()
			t.writeToStdout([]byte("exit\n"))
			return true
		}
	}

	// Send everything as stdin to remote terminal
	msg := ws.NewMessage(ws.MessageTypeStdin, data, t.wsClient.GetMicroserviceUUID(), t.wsClient.GetExecID())
	t.wsClient.SendMessage(msg)
	return false
}

func (t *Terminal) redrawInputLine() {
	t.stdoutMutex.Lock()
	defer t.stdoutMutex.Unlock()

	// Clear the current line and move to start
	os.Stdout.Write([]byte("\r\x1b[K"))

	// Redraw the prompt and input
	if t.prompt != "" {
		os.Stdout.Write([]byte(t.prompt))
	}
	os.Stdout.Write([]byte(string(t.inputBuffer)))

	// Move cursor to end of input
	t.cursorPos = len(t.inputBuffer)
	os.Stdout.Write([]byte("\r")) // Move to start of line first
	if t.prompt != "" {
		os.Stdout.Write([]byte(t.prompt)) // Move past prompt
	}
	if t.cursorPos > 0 {
		os.Stdout.Write([]byte(fmt.Sprintf("\x1b[%dC", t.cursorPos))) // Move to cursor position
	}
	os.Stdout.Sync()
}

func (t *Terminal) cleanup() {
	t.cleanupOnce.Do(func() {
		if t.wsClient != nil {
			t.wsClient.Close()
		}
		if t.oldState != nil {
			term.Restore(int(os.Stdin.Fd()), t.oldState)
			t.oldState = nil
		}
	})
}

func (t *Terminal) Start() error {
	var err error
	t.oldState, err = term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to set terminal to raw mode: %v", err)
	}
	defer t.cleanup()

	// Windows doesn't support SIGWINCH, so we skip resize handling
	// signal.Notify(t.resizeCh, syscall.SIGWINCH)
	// go t.handleResize()

	// Create error channel to coordinate exit
	errCh := make(chan error, 1)

	// Monitor output and errors
	go func() {
		for {
			select {
			case <-t.ctx.Done():
				return
			default:
				msg, err := t.wsClient.ReadMessage()
				if err != nil {
					// Check if this is a normal closure
					if t.wsClient.IsNormalClosure(err) {
						// Normal closure - don't report as error, just exit gracefully
						t.cancel()
						return
					}
					// This is an actual error - pass it to exec layer for handling
					errCh <- err
					t.cancel()
					return
				}
				if msg == nil {
					// Normal termination (no error, no message)
					t.cancel()
					return
				}
				output := string(msg.Data)
				t.writeToStdout([]byte(output))
			}
		}
	}()

	// Check for initial WebSocket errors
	if err := t.wsClient.GetError(); err != nil {
		// Check if this is a normal closure
		if t.wsClient.IsNormalClosure(err) {
			// Normal closure - don't report as error
			t.cancel()
			return nil
		}
		// This is an actual error - pass it to exec layer for handling
		t.cancel()
		return err
	}

	// Start input handling in a separate goroutine
	inputDone := make(chan struct{})
	go func() {
		defer close(inputDone)
		buf := make([]byte, 1)
		for {
			select {
			case <-t.ctx.Done():
				return
			default:
				n, err := os.Stdin.Read(buf)
				if err != nil || n == 0 {
					continue
				}

				// Handle input
				if t.handleInput(buf[:n]) {
					t.cancel()
					return
				}
			}
		}
	}()

	// Wait for either error or context cancellation
	select {
	case <-t.ctx.Done():
		return nil
	case err := <-errCh:
		t.cancel()
		return err
	case <-inputDone:
		return nil
	}
}

// func (t *Terminal) handleResize() {
// 	// Windows doesn't support SIGWINCH, so this function is empty
// 	for range t.resizeCh {
// 		// No-op on Windows
// 	}
// }

func (t *Terminal) Stop() {
	t.cancel()
	t.cleanup()
}
