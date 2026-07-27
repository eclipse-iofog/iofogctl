package install

// Engine restart state machine (P9-HYG-3). See 09-edgelet-install-hygiene.md §8.
//
// | Trigger                         | Who restarts containerd      | Wait for socket |
// |---------------------------------|------------------------------|-----------------|
// | Fresh install + WASM shims      | First start_edgelet only     | Go poll         |
// | Redeploy + WASM shim change     | Go restartDeferredWasmEngine | Go poll         |
// | Redeploy + embed OTA hash change| Shell marker → Go deferred   | Go poll         |
// | Redeploy, no engine change      | Skip                         | Go poll only    |
//
// Shell markers: EDGELET_WASM_SKIP_RESTART, EDGELET_WASM_DEFER_ENGINE_RESTART,
// .restart-data-plane (lib/embed.sh), .containerd-restarted (lib/service.sh).

// HostInstallState captures host facts used to plan an engine restart.
type HostInstallState struct {
	EngineActive bool
	WasmChanged  bool
	EmbedOTA     bool
}

// EngineRestartPlan describes whether and when to restart edgelet-containerd.
type EngineRestartPlan struct {
	Needed   bool
	Deferred bool
	Reason   string
}

// PlanEngineRestart decides containerd restart timing from install config and host state.
func PlanEngineRestart(cfg EdgeletInstallConfig, prior *HostInstallState) EngineRestartPlan {
	if cfg.containerEngine() != "edgelet" {
		return EngineRestartPlan{}
	}
	state := HostInstallState{}
	if prior != nil {
		state = *prior
	}
	if !state.EngineActive {
		state.EngineActive = cfg.engineActive
	}

	if state.EmbedOTA && state.EngineActive {
		return EngineRestartPlan{Needed: true, Deferred: true, Reason: "embed-ota"}
	}
	if state.WasmChanged && state.EngineActive {
		return EngineRestartPlan{Needed: true, Deferred: true, Reason: "wasm-changed"}
	}
	return EngineRestartPlan{}
}

func hostStateFromWasm(freshInstall bool, anyChanged bool) *HostInstallState {
	return &HostInstallState{
		EngineActive: !freshInstall,
		WasmChanged:  anyChanged,
	}
}
