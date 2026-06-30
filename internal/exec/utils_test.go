package exec

import (
	"errors"
	"fmt"
	"testing"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
)

func TestFormatExecError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "relay unavailable",
			err:  fmt.Errorf("%w: Relay unavailable for cross-replica session", client.ErrWsRelayUnavailable),
			want: ErrMsgExecRelayUnavailable,
		},
		{
			name: "legacy router reason",
			err:  fmt.Errorf("%w: Router unavailable for cross-replica session", client.ErrWsRelayUnavailable),
			want: ErrMsgExecRelayUnavailable,
		},
		{
			name: "agent timeout",
			err:  fmt.Errorf("%w: Timeout waiting for agent connection", client.ErrWsAgentTimeout),
			want: ErrMsgTimeoutWaitingForAgent,
		},
		{
			name: "server draining",
			err:  fmt.Errorf("%w: Server draining", client.ErrWsServerDraining),
			want: ErrMsgExecServerDraining,
		},
		{
			name: "session quota",
			err:  fmt.Errorf("%w: Maximum of 3 concurrent exec sessions allowed", client.ErrExecSessionQuotaExceeded),
			want: ErrMsgExecSessionQuotaExceeded,
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
			got := formatExecError(tt.err)
			if got != tt.want {
				t.Fatalf("formatExecError() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatExecErrorRelayNotAgentTimeout(t *testing.T) {
	t.Parallel()

	relayErr := fmt.Errorf("%w: Relay unavailable for cross-replica session", client.ErrWsRelayUnavailable)
	if errors.Is(relayErr, client.ErrWsAgentTimeout) {
		t.Fatal("relay error must not match agent timeout")
	}
	if got := formatExecError(relayErr); got == ErrMsgTimeoutWaitingForAgent {
		t.Fatalf("relay close must not map to agent timeout message: %q", got)
	}
}
