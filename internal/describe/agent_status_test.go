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
		CPUUsage:              150,
		DiskUsage:             0.000000187, // 187 B as GiB (decimal) from Edgelet status PUT
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

	got := statusMap(FormatAgentStatus(status))

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
	got := statusMap(FormatAgentStatus(rsc.AgentStatus{}))

	require.Contains(t, got, "availableRuntimes")
	require.Contains(t, got, "runtimeAgentPhase")
	require.Contains(t, got, "controlPlaneQuiesced")
	require.Equal(t, "", got["runtimeAgentPhase"])
	require.Equal(t, false, got["controlPlaneQuiesced"])
}

func TestFormatAgentStatusDiskUsageGiB(t *testing.T) {
	got := statusMap(FormatAgentStatus(rsc.AgentStatus{DiskUsage: 0.24}))

	require.Equal(t, "228.9 MB", got["diskUsage"])
}

func TestFormatAgentStatusUptimeAndTimestamps(t *testing.T) {
	// 2026-05-07T15:21:10+03:00 in ms since epoch
	tsMilli := time.Date(2026, 5, 7, 15, 21, 10, 0, time.FixedZone("TRT", 3*3600)).UnixMilli()
	status := rsc.AgentStatus{
		LastActive:          tsMilli,
		LastStatusTimeMsUTC: tsMilli,
		UptimeMs:            (5*time.Hour + 3*time.Minute).Milliseconds(),
	}

	got := statusMap(FormatAgentStatus(status))

	require.Equal(t, "2026-05-07T12:21:10Z", got["lastActive"])
	require.Equal(t, "2026-05-07T12:21:10Z", got["lastStatusTime"])
	require.Equal(t, "5h3m", got["uptime"])
}

func TestFormatAgentStatusV39ModelAndRuntimeBlobs(t *testing.T) {
	got := statusMap(FormatAgentStatus(rsc.AgentStatus{
		RuntimeClasses:      `[{"name":"nvidia"}]`,
		AvailableCdiDevices: `["nvidia.com/gpu=0"]`,
		ModelStatus:         `[{"name":"llama","state":"Ready"}]`,
		ActiveModels:        2,
		ModelLastUpdate:     1710000000000,
	}))

	require.Equal(t, 2, got["activeModels"])
	require.Equal(t, "2024-03-09T16:00:00Z", got["modelLastUpdate"])

	runtimeClasses, ok := got["runtimeClasses"].([]interface{})
	require.True(t, ok)
	require.Len(t, runtimeClasses, 1)
	require.Equal(t, "nvidia", runtimeClasses[0].(map[string]interface{})["name"])

	cdi, ok := got["availableCdiDevices"].([]interface{})
	require.True(t, ok)
	require.Equal(t, "nvidia.com/gpu=0", cdi[0])

	models, ok := got["modelStatus"].([]interface{})
	require.True(t, ok)
	require.Equal(t, "llama", models[0].(map[string]interface{})["name"])
}

func TestFormatAgentStatusV39KnowledgeFields(t *testing.T) {
	got := statusMap(FormatAgentStatus(rsc.AgentStatus{
		KnowledgeStatus:     `[{"name":"wiki"}]`,
		ActiveKnowledge:     3,
		KnowledgeLastUpdate: 1710000000456,
	}))

	require.Equal(t, 3, got["activeKnowledge"])
	require.Equal(t, "2024-03-09T16:00:00Z", got["knowledgeLastUpdate"])

	knowledge, ok := got["knowledgeStatus"].([]interface{})
	require.True(t, ok)
	require.Equal(t, "wiki", knowledge[0].(map[string]interface{})["name"])
}

func TestFormatAgentStatusOmitsEmptyV39Blobs(t *testing.T) {
	got := statusMap(FormatAgentStatus(rsc.AgentStatus{}))
	require.NotContains(t, got, "runtimeClasses")
	require.NotContains(t, got, "availableCdiDevices")
	require.NotContains(t, got, "modelStatus")
	require.NotContains(t, got, "modelLastUpdate")
	require.NotContains(t, got, "knowledgeStatus")
	require.NotContains(t, got, "knowledgeLastUpdate")
	require.NotContains(t, got, "systemCpus")
	require.NotContains(t, got, "systemTotalMemory")
	require.NotContains(t, got, "systemAvailableMemory")
	require.NotContains(t, got, "systemTotalDisk")
	require.NotContains(t, got, "systemAvailableDisk")
	require.NotContains(t, got, "systemTotalCPU")
	require.NotContains(t, got, "systemOs")
	require.NotContains(t, got, "processedMessages")
	require.NotContains(t, got, "messageSpeed")
	require.Equal(t, 0, got["activeModels"])
	require.Equal(t, 0, got["activeKnowledge"])
}

func TestFormatAgentStatusHostMetrics(t *testing.T) {
	got := statusMap(FormatAgentStatus(rsc.AgentStatus{
		CPUUsage:              32.97,
		SystemCpus:            4,
		SystemTotalMemory:     8 * 1024 * 1024 * 1024,
		SystemAvailableMemory: 3 * 1024 * 1024 * 1024,
		SystemTotalDisk:       50 * 1024 * 1024 * 1024,
		SystemAvailableDisk:   4 * 1024 * 1024 * 1024,
		SystemTotalCPU:        21.03,
		SystemOs:              "linux",
		SystemOsVersion:       "6.8.0",
		SystemKernelVersion:   "6.8.0-60-generic",
	}))

	require.Equal(t, "0.33 cores", got["cpuUsage"])
	require.Equal(t, 4, got["systemCpus"])
	require.Equal(t, "8.0 GB", got["systemTotalMemory"])
	require.Equal(t, "3.0 GB", got["systemAvailableMemory"])
	require.Equal(t, "50.0 GB", got["systemTotalDisk"])
	require.Equal(t, "4.0 GB", got["systemAvailableDisk"])
	require.Equal(t, "21.03 %", got["systemTotalCPU"])
	require.Equal(t, "linux", got["systemOs"])
	require.Equal(t, "6.8.0", got["systemOsVersion"])
	require.Equal(t, "6.8.0-60-generic", got["systemKernelVersion"])
}

func TestFormatAgentStatusFieldOrder(t *testing.T) {
	keys := statusKeys(FormatAgentStatus(rsc.AgentStatus{
		Version:               "v1.1.0-rc.5",
		DaemonStatus:          "RUNNING",
		UptimeMs:              time.Hour.Milliseconds(),
		CPUUsage:              150,
		MemoryUsage:           152,
		DiskUsage:             0.24,
		SystemOs:              "linux",
		SystemCpus:            4,
		SystemTotalCPU:        21.03,
		SystemTotalMemory:     8 * 1024 * 1024 * 1024,
		SystemAvailableMemory: 3 * 1024 * 1024 * 1024,
		ActiveModels:          1,
		ActiveKnowledge:       2,
	}))

	require.Equal(t, []string{
		"version",
		"daemonStatus",
		"securityStatus",
		"securityViolationInfo",
		"warningMessage",
		"gpsStatus",
		"ipAddress",
		"ipAddressExternal",
		"lastCommandTime",
		"uptime",
		"cpuUsage",
		"memoryUsage",
		"diskUsage",
		"cpuViolation",
		"memoryViolation",
		"diskViolation",
		"systemOs",
		"systemCpus",
		"systemTotalCPU",
		"systemTotalMemory",
		"systemAvailableMemory",
		"availableRuntimes",
		"runtimeAgentPhase",
		"controlPlaneQuiesced",
		"activeModels",
		"activeKnowledge",
		"repositoryStatus",
		"isReadyToUpgrade",
		"isReadyToRollback",
		"tunnel",
		"volumeMounts",
	}, keys)
}
