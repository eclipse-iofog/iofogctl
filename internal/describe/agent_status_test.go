package describe

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func agentStatusFixturePath(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata", name)
}

func loadAgentStatusFixture(t *testing.T, name string) map[string]interface{} {
	t.Helper()
	data, err := os.ReadFile(agentStatusFixturePath(name))
	require.NoError(t, err)
	var expected map[string]interface{}
	require.NoError(t, yaml.Unmarshal(data, &expected))
	return expected
}

func TestFormatAgentStatusGoldenV38Fields(t *testing.T) {
	expected := loadAgentStatusFixture(t, "agent-config-status-v38.yaml")

	status := rsc.AgentStatus{
		DaemonStatus:          "UNKNOWN",
		SecurityStatus:        "OK",
		WarningMessage:        "HEALTHY",
		SecurityViolationInfo: "No violation",
		CPUUsage:              0.13,
		DiskUsage:             0.00017833709716796875, // ~187 B as MiB input
		MemoryViolation:       "false",
		DiskViolation:         "false",
		CPUViolation:          "false",
		RepositoryStatus:      "[]",
		IPAddress:             "172.20.0.1",
		IPAddressExternal:     "176.234.134.248",
		LastCommandTimeMsUTC:  0,
		Version:               "3.7.0",
		IsReadyToUpgrade:      false,
		IsReadyToRollback:     false,
		Tunnel:                "",
		GpsStatus:             "HEALTHY",
		AvailableRuntimes:     []string{"crun", "x-runtime", "y-runtime"},
		RuntimeAgentPhase:     "",
		ControlPlaneQuiesced:  false,
	}

	got := FormatAgentStatus(status)

	for key, want := range expected {
		require.Contains(t, got, key, "missing status field %q", key)
		if key == "availableRuntimes" {
			require.Equal(t, []string{"crun", "x-runtime", "y-runtime"}, got[key])
			continue
		}
		require.Equal(t, want, got[key], "status field %q", key)
	}
}

func TestFormatAgentStatusV38FieldsAlwaysPresent(t *testing.T) {
	got := FormatAgentStatus(rsc.AgentStatus{})

	require.Contains(t, got, "availableRuntimes")
	require.Contains(t, got, "runtimeAgentPhase")
	require.Contains(t, got, "controlPlaneQuiesced")
	require.Equal(t, "", got["runtimeAgentPhase"])
	require.Equal(t, false, got["controlPlaneQuiesced"])
}

func TestFormatAgentStatusUptimeAndTimestamps(t *testing.T) {
	// 2026-05-07T15:21:10+03:00 in ms since epoch
	tsMilli := time.Date(2026, 5, 7, 15, 21, 10, 0, time.FixedZone("TRT", 3*3600)).UnixMilli()
	status := rsc.AgentStatus{
		LastActive:          tsMilli,
		LastStatusTimeMsUTC: tsMilli,
		UptimeMs:            (5*time.Hour + 3*time.Minute).Milliseconds(),
	}

	got := FormatAgentStatus(status)

	require.Equal(t, "2026-05-07T15:21:10+03:00", got["lastActive"])
	require.Equal(t, "2026-05-07T15:21:10+03:00", got["lastStatusTime"])
	require.Equal(t, "5h3m", got["uptime"])
}
