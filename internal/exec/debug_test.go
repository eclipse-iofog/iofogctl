package exec

import (
	"testing"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
)

func TestIsDebugMicroserviceName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		msvcName  string
		agentName string
		agentUUID string
		want      bool
	}{
		{name: "canonical debug", msvcName: "debug", agentName: "remote-1", agentUUID: "node-remote-1", want: true},
		{name: "legacy debug uuid", msvcName: "debug-node-legacy", agentName: "legacy-agent", agentUUID: "node-legacy", want: true},
		{name: "debug agent name", msvcName: "debug-edge-1", agentName: "edge-1", agentUUID: "uuid-1", want: true},
		{name: "router", msvcName: "router", agentName: "remote-1", agentUUID: "node-remote-1", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isDebugMicroserviceName(tt.msvcName, tt.agentName, tt.agentUUID); got != tt.want {
				t.Fatalf("isDebugMicroserviceName(%q) = %v, want %v", tt.msvcName, got, tt.want)
			}
		})
	}
}

func TestIsDebugExecAlreadyProvisioned(t *testing.T) {
	t.Parallel()

	if !isDebugExecAlreadyProvisioned(client.NewConflictError("debug exec already exists")) {
		t.Fatal("expected conflict attach error to be treated as already provisioned")
	}
	if isDebugExecAlreadyProvisioned(errString("permission denied")) {
		t.Fatal("expected unrelated error to remain fatal")
	}
}

type errString string

func (e errString) Error() string { return string(e) }
