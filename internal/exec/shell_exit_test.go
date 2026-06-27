package exec

import (
	"fmt"
	"testing"
)

func TestIsShellExitInput(t *testing.T) {
	tests := []struct {
		name       string
		data       string
		lineBuffer string
		want       bool
	}{
		{"exit command", "exit\r", "exit", true},
		{"logout command", "logout\n", "logout", true},
		{"ctrl+d", "\x04", "", true},
		{"partial exit", "ex", "ex", false},
		{"other command", "ls\r", "ls", false},
		{"exit with spaces", "exit\r", "  exit  ", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isShellExitInput(tt.data, tt.lineBuffer); got != tt.want {
				t.Fatalf("isShellExitInput(%q, %q) = %v, want %v", tt.data, tt.lineBuffer, got, tt.want)
			}
		})
	}
}

func TestIsNormalExecClose(t *testing.T) {
	err1005 := fmt.Errorf("%s: ", execCloseCodeString(1005))
	err1000 := fmt.Errorf("%s: done", execCloseCodeString(1000))

	if !isNormalExecClose(err1000, false) {
		t.Fatal("expected code 1000 to be normal")
	}
	if isNormalExecClose(err1005, false) {
		t.Fatal("expected code 1005 without user close to be abnormal")
	}
	if !isNormalExecClose(err1005, true) {
		t.Fatal("expected code 1005 with user close to be normal")
	}
}
