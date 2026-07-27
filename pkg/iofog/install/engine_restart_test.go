package install

import "testing"

func TestPlanEngineRestart_FreshInstallSkips(t *testing.T) {
	cfg := EdgeletInstallConfig{ContainerEngine: "edgelet"}
	plan := PlanEngineRestart(cfg, hostStateFromWasm(true, true))
	if plan.Needed || plan.Deferred {
		t.Fatalf("fresh install should skip engine restart, got %+v", plan)
	}
}

func TestPlanEngineRestart_WasmRedeployDefers(t *testing.T) {
	cfg := EdgeletInstallConfig{ContainerEngine: "edgelet", engineActive: true}
	plan := PlanEngineRestart(cfg, hostStateFromWasm(false, true))
	if !plan.Needed || !plan.Deferred || plan.Reason != "wasm-changed" {
		t.Fatalf("redeploy wasm change should defer restart, got %+v", plan)
	}
}

func TestPlanEngineRestart_EmbedOTADefers(t *testing.T) {
	cfg := EdgeletInstallConfig{ContainerEngine: "edgelet", engineActive: true}
	plan := PlanEngineRestart(cfg, &HostInstallState{EngineActive: true, EmbedOTA: true})
	if !plan.Needed || !plan.Deferred || plan.Reason != "embed-ota" {
		t.Fatalf("embed OTA should defer restart, got %+v", plan)
	}
}

func TestPlanEngineRestart_NonEdgeletEngine(t *testing.T) {
	cfg := EdgeletInstallConfig{ContainerEngine: "docker", engineActive: true}
	plan := PlanEngineRestart(cfg, hostStateFromWasm(false, true))
	if plan.Needed || plan.Deferred {
		t.Fatalf("docker engine should skip edgelet-containerd restart, got %+v", plan)
	}
}

func TestPlanEngineRestart_UnchangedRedeploy(t *testing.T) {
	cfg := EdgeletInstallConfig{ContainerEngine: "edgelet", engineActive: true}
	plan := PlanEngineRestart(cfg, hostStateFromWasm(false, false))
	if plan.Needed || plan.Deferred {
		t.Fatalf("unchanged redeploy should skip restart, got %+v", plan)
	}
}
