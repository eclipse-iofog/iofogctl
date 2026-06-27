package exec

import (
	"fmt"
	"strconv"
	"strings"
)

func isShellExitInput(data, lineBuffer string) bool {
	if strings.ContainsRune(data, '\x04') {
		return true
	}
	if !strings.ContainsAny(data, "\r\n") {
		return false
	}
	line := strings.TrimSpace(strings.NewReplacer("\r", "", "\n", "").Replace(lineBuffer))
	return line == "exit" || line == "logout"
}

func parseExecWSCloseCode(err error) (int, bool) {
	if err == nil {
		return 0, false
	}
	const prefix = "exec WebSocket closed (code "
	msg := err.Error()
	idx := strings.Index(msg, prefix)
	if idx < 0 {
		return 0, false
	}
	rest := msg[idx+len(prefix):]
	end := strings.IndexByte(rest, ')')
	if end < 0 {
		return 0, false
	}
	code, convErr := strconv.Atoi(rest[:end])
	if convErr != nil {
		return 0, false
	}
	return code, true
}

func isNormalExecClose(err error, userInitiatedClose bool) bool {
	if err == nil {
		return true
	}

	if code, ok := parseExecWSCloseCode(err); ok {
		if code == 1000 {
			return true
		}
		if userInitiatedClose && code == 1005 {
			return true
		}
	}

	errStr := err.Error()
	if userInitiatedClose && strings.HasPrefix(errStr, "exec WebSocket closed (code") {
		return true
	}
	if strings.Contains(errStr, "use of closed network connection") {
		return true
	}

	return false
}

func execCloseCodeString(code int) string {
	return fmt.Sprintf("exec WebSocket closed (code %d)", code)
}
