package describe

import (
	"testing"
	"time"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/stretchr/testify/require"
)

func TestFormatAgentStatusPlatformStatus(t *testing.T) {
	lastErr := "router reconcile timeout"
	transition := time.Date(2026, 6, 24, 22, 52, 59, 824000000, time.UTC)
	status := rsc.AgentStatus{
		DaemonStatus:   "RUNNING",
		WarningMessage: "Platform reconcile: progressing",
		PlatformStatus: &client.PlatformStatus{
			Phase:              client.PlatformReady,
			Generation:         2,
			ObservedGeneration: 2,
			LastError:          &lastErr,
			LastTransitionAt:   &transition,
			Conditions: []client.PlatformCondition{
				{Type: "RouterReady", Status: "True", Reason: "ReconcileComplete"},
				{Type: "NatsReady", Status: "True", Reason: "ReconcileComplete"},
			},
		},
	}

	got := FormatAgentStatus(status)
	ps, ok := got["platformStatus"].(map[string]interface{})
	require.True(t, ok, "platformStatus should be present")
	require.Equal(t, "Ready", ps["phase"])
	require.Equal(t, 2, ps["generation"])
	require.Equal(t, 2, ps["observedGeneration"])
	require.Equal(t, lastErr, ps["lastError"])
	require.Equal(t, transition.Format(time.RFC3339Nano), ps["lastTransitionAt"])

	conditions, ok := ps["conditions"].([]map[string]interface{})
	require.True(t, ok)
	require.Len(t, conditions, 2)
	require.Equal(t, "RouterReady", conditions[0]["type"])
}

func TestFormatAgentStatusOmitsNilPlatformStatus(t *testing.T) {
	got := FormatAgentStatus(rsc.AgentStatus{})
	require.NotContains(t, got, "platformStatus")
}
