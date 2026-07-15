package util

import (
	"errors"
	"testing"
)

func TestIsSSHSignalTerminated(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"exit 143", errors.New("Process exited with status 143"), true},
		{"signal killed", errors.New("signal: killed"), true},
		{"other", errors.New("exit status 1"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsSSHSignalTerminated(tc.err); got != tc.want {
				t.Fatalf("IsSSHSignalTerminated(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}
