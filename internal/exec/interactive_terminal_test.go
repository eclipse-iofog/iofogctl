package exec

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
)

type fakeExecSession struct {
	stdin    []byte
	control  []byte
	writeErr error
}

func (f *fakeExecSession) Read() (*client.ExecFrame, error) { return nil, errors.New("unused") }
func (f *fakeExecSession) Close() error                     { return nil }
func (f *fakeExecSession) WriteStdin(data []byte) error {
	if f.writeErr != nil {
		return f.writeErr
	}
	f.stdin = append(f.stdin, data...)
	return nil
}
func (f *fakeExecSession) WriteControl(data []byte) error {
	if f.writeErr != nil {
		return f.writeErr
	}
	f.control = append([]byte(nil), data...)
	return nil
}

func TestHandleInputForwardsCtrlC(t *testing.T) {
	session := &fakeExecSession{}
	term := newInteractiveTerminal(session)

	if term.handleInput([]byte{0x03}) {
		t.Fatal("Ctrl+C must be forwarded and must not end the relay")
	}
	if !bytes.Equal(session.stdin, []byte{0x03}) {
		t.Fatalf("stdin = %v, want Ctrl+C byte", session.stdin)
	}
	if len(session.control) != 0 {
		t.Fatalf("control = %s, want no local control frame", session.control)
	}
}

func TestHandleInputForwardsChunk(t *testing.T) {
	session := &fakeExecSession{}
	term := newInteractiveTerminal(session)
	input := []byte("python3 rest_infer.py")
	if term.handleInput(input) {
		t.Fatal("input write must not end the relay")
	}
	if !bytes.Equal(session.stdin, input) {
		t.Fatalf("stdin = %q, want %q", session.stdin, input)
	}
}

func TestResizeControlPayload(t *testing.T) {
	payload, err := resizeControlPayload(120, 40)
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Type string `json:"type"`
		Cols uint32 `json:"cols"`
		Rows uint32 `json:"rows"`
	}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Type != "resize" || decoded.Cols != 120 || decoded.Rows != 40 {
		t.Fatalf("payload = %+v", decoded)
	}

	if _, err := resizeControlPayload(0, 24); err == nil {
		t.Fatal("zero columns must be rejected")
	}
	if _, err := resizeControlPayload(10, maxExecTermDimension+1); err == nil {
		t.Fatal("oversized rows must be rejected")
	}
}

func TestWriteResizeSendsControlFrame(t *testing.T) {
	session := &fakeExecSession{}
	term := newInteractiveTerminal(session)
	if err := term.writeResize(100, 30); err != nil {
		t.Fatal(err)
	}
	payload, err := resizeControlPayload(100, 30)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(session.control, payload) {
		t.Fatalf("control = %s, want %s", session.control, payload)
	}
	if err := term.writeResize(0, 24); err == nil {
		t.Fatal("invalid size must not be sent")
	}
}
