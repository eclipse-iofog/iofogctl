package util

import (
	"io"
	"os"
)

// IgnoreErr discards an error (interactive I/O cleanup paths).
func IgnoreErr(err error) {}

// IgnoreClose closes c and discards errors.
func IgnoreClose(c io.Closer) {
	if c != nil {
		_ = c.Close()
	}
}

// WriteStdout writes data to stdout and syncs when possible.
func WriteStdout(data []byte) {
	if _, err := os.Stdout.Write(data); err != nil {
		return
	}
	_ = os.Stdout.Sync()
}

// WriteStdoutString writes s to stdout and syncs when possible.
func WriteStdoutString(s string) {
	WriteStdout([]byte(s))
}

// WriteStderrString writes s to stderr.
func WriteStderrString(s string) {
	_, _ = os.Stderr.WriteString(s)
}

// DrainAndCloseHTTPBody drains and closes an HTTP response body.
func DrainAndCloseHTTPBody(body io.ReadCloser) {
	if body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, body)
	_ = body.Close()
}
