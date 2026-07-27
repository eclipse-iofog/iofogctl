package install

import (
	"context"
	"fmt"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func (agent *RemoteEdgelet) restartDeferredWasmEngine() error {
	plan := PlanEngineRestart(agent.cfg, hostStateFromWasm(!agent.cfg.engineActive, agent.cfg.wasmDeferEngineRestart))
	if !plan.Deferred || !plan.Needed {
		return nil
	}

	util.PrintInfo("Restarting edgelet-containerd for WASM shim update (drain may take up to 120s)")

	restartCmd := wrapBootstrapCommand(restartEdgeletContainerdCommand(agent.dir), agent.remoteBootstrapEnv(), true)
	if err := agent.run([]command{{
		cmd: restartCmd,
		msg: "Restarting edgelet-containerd on " + agent.name,
	}}); err != nil {
		if !util.IsSSHSignalTerminated(err) {
			return err
		}
		if ready, checkErr := agent.remoteContainerdReady(); checkErr != nil {
			return err
		} else if !ready {
			return err
		}
		util.PrintInfo("edgelet-containerd restart already in progress after interrupted SSH session")
	}

	if err := agent.pollContainerdReady(context.Background()); err != nil {
		return err
	}

	if err := agent.run([]command{{
		cmd: markContainerdRestartedShell(agent.dir),
		msg: "Marking edgelet-containerd restarted on " + agent.name,
	}}); err != nil {
		return fmt.Errorf("mark containerd restarted: %w", err)
	}
	return nil
}

func (agent *RemoteEdgelet) remoteContainerdReady() (bool, error) {
	if remoteEdgeletRunHook != nil {
		return true, nil
	}
	if err := agent.ssh.Connect(); err != nil {
		return false, err
	}
	defer util.Log(agent.ssh.Disconnect)
	_, err := agent.ssh.Run(agent.remoteProbeCommand(pkg.edgeletScriptProbeContainerdReady))
	return err == nil, nil
}
