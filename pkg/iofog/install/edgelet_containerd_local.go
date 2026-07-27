package install

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func (agent *LocalEdgelet) restartDeferredWasmEngine() error {
	plan := PlanEngineRestart(agent.cfg, hostStateFromWasm(!agent.cfg.engineActive, agent.cfg.wasmDeferEngineRestart))
	if !plan.Deferred || !plan.Needed {
		return nil
	}

	util.PrintInfo("Restarting edgelet-containerd for WASM shim update (drain may take up to 120s)")
	cmd := wrapBootstrapCommand(restartEdgeletContainerdCommand(agent.dir), agent.cfg.bootstrapEnv(true), true)
	if _, err := util.Exec("", "sh", "-c", cmd); err != nil {
		if !util.IsSSHSignalTerminated(err) {
			return err
		}
	}

	return pollContainerdReadyLocal(agent.dir)
}

func (agent *LocalEdgelet) markContainerdRestartedLocal() error {
	_, err := util.Exec("", "sh", "-c", markContainerdRestartedShell(agent.dir))
	return err
}

// finalizeDeferredWasmEngine marks containerd restarted after local wait (remote marks in restartDeferredWasmEngine).
func (agent *LocalEdgelet) finalizeDeferredWasmEngine() error {
	if !agent.cfg.wasmDeferEngineRestart || agent.cfg.containerEngine() != "edgelet" {
		return nil
	}
	if err := agent.markContainerdRestartedLocal(); err != nil {
		return fmt.Errorf("mark containerd restarted: %w", err)
	}
	return nil
}
