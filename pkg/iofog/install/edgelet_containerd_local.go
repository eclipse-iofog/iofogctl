package install

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func (agent *LocalEdgelet) restartDeferredWasmEngine() error {
	if !agent.cfg.wasmDeferEngineRestart || agent.cfg.containerEngine() != "edgelet" {
		return nil
	}

	util.PrintInfo("Restarting edgelet-containerd for WASM shim update (drain may take up to 120s)")
	if _, err := util.Exec("", "sh", "-c", restartEdgeletContainerdNoBlockShell()); err != nil {
		if !util.IsSSHSignalTerminated(err) {
			return err
		}
	}

	return waitForEdgeletContainerdReady(func() (bool, error) {
		_, err := util.Exec("", "sh", "-c", edgeletContainerdReadyCheckShell())
		return err == nil, nil
	})
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
