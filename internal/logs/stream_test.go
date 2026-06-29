package logs

import (
	"errors"
	"fmt"
	"testing"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
)

func TestFormatLogError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "relay unavailable",
			err:  fmt.Errorf("%w: Relay unavailable for cross-replica session", client.ErrWsRelayUnavailable),
			want: errMsgLogRelayUnavailable,
		},
		{
			name: "legacy router reason",
			err:  fmt.Errorf("%w: Router unavailable for cross-replica session", client.ErrWsRelayUnavailable),
			want: errMsgLogRelayUnavailable,
		},
		{
			name: "agent timeout",
			err:  fmt.Errorf("%w: Timeout waiting for agent connection", client.ErrWsAgentTimeout),
			want: errMsgTimeoutWaitingForAgent,
		},
		{
			name: "server draining",
			err:  fmt.Errorf("%w: Server draining", client.ErrWsServerDraining),
			want: errMsgLogServerDraining,
		},
		{
			name: "log session unavailable",
			err:  fmt.Errorf("%w: No available log session", client.ErrLogSessionUnavailable),
			want: client.ErrLogSessionUnavailable.Error(),
		},
		{
			name: "nil",
			err:  nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := formatLogError(tt.err)
			if got != tt.want {
				t.Fatalf("formatLogError() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatLogErrorRelayNotPolicyViolation(t *testing.T) {
	t.Parallel()

	relayErr := fmt.Errorf("%w: Relay unavailable for cross-replica session", client.ErrWsRelayUnavailable)
	if errors.Is(relayErr, client.ErrLogPolicyViolation) {
		t.Fatal("relay error must not match log policy violation")
	}
	if got := formatLogError(relayErr); got == "Policy violation: Access denied" {
		t.Fatalf("relay close must not map to generic policy violation: %q", got)
	}
}
