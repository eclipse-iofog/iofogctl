package install

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func (agent *RemoteEdgelet) restartDeferredWasmEngine() error {
	if !agent.cfg.wasmDeferEngineRestart || agent.cfg.containerEngine() != "edgelet" {
		return nil
	}
	if remoteEdgeletRunHook != nil {
		return nil
	}

	util.PrintInfo("Restarting edgelet-containerd for WASM shim update (drain may take up to 120s)")

	stream := util.SSHRunOptions{StreamOutput: true}
	if err := agent.ssh.Connect(); err != nil {
		return err
	}
	defer util.Log(agent.ssh.Disconnect)

	if _, err := agent.ssh.RunWithOptions(restartEdgeletContainerdNoBlockShell(), stream); err != nil {
		if !util.IsSSHSignalTerminated(err) {
			return err
		}
		if ready, checkErr := agent.remoteEdgeletContainerdReady(); checkErr != nil {
			return err
		} else if !ready {
			return err
		}
		util.PrintInfo("edgelet-containerd restart already in progress after interrupted SSH session")
	}

	if err := waitForEdgeletContainerdReady(agent.remoteEdgeletContainerdReady); err != nil {
		return err
	}

	if _, err := agent.ssh.RunWithOptions(markContainerdRestartedShell(agent.dir), stream); err != nil {
		return fmt.Errorf("mark containerd restarted: %w", err)
	}
	return nil
}

func (agent *RemoteEdgelet) remoteEdgeletContainerdReady() (bool, error) {
	if remoteEdgeletRunHook != nil {
		return true, nil
	}
	if err := agent.ssh.Connect(); err != nil {
		return false, err
	}
	_, err := agent.ssh.Run(edgeletContainerdReadyCheckShell())
	return err == nil, nil
}
