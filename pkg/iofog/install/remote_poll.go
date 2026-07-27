package install

import (
	"context"
	"fmt"
	"time"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type pollOptions struct {
	interval time.Duration
	timeout  time.Duration
	label    string
}

const verifyProbeTimeout = 30 * time.Second

// remotePollScriptHook is set by tests to mock probe script execution over SSH.
var remotePollScriptHook func(agent *RemoteEdgelet, script string) error

// remoteVerifyHostHook is set by tests to mock host verification after SSH 143.
var remoteVerifyHostHook func(agent *RemoteEdgelet) (healthy bool, err error)

func (agent *RemoteEdgelet) remoteBootstrapEnv() string {
	return agent.cfg.bootstrapEnv(false)
}

func (agent *RemoteEdgelet) remoteProbeCommand(script string) string {
	scriptPath := util.JoinAgentPath(agent.dir, script)
	return wrapBootstrapCommand("sudo "+scriptPath, agent.remoteBootstrapEnv(), true)
}

func (agent *RemoteEdgelet) runRemoteProbe(script string) error {
	if remotePollScriptHook != nil {
		return remotePollScriptHook(agent, script)
	}
	if remoteEdgeletRunHook != nil {
		return nil
	}

	if err := agent.ssh.Disconnect(); err != nil {
		util.SSHVerbose(fmt.Sprintf("probe disconnect before %s: %v", script, err))
	}
	if err := agent.ssh.Connect(); err != nil {
		return err
	}
	defer util.Log(agent.ssh.Disconnect)

	_, err := agent.ssh.Run(agent.remoteProbeCommand(script))
	return err
}

func (agent *RemoteEdgelet) pollRemoteScript(ctx context.Context, script string, opts pollOptions) error {
	if remoteEdgeletRunHook != nil && remotePollScriptHook == nil {
		return nil
	}

	timeoutSec := int(opts.timeout.Seconds())
	util.PrintInfo(fmt.Sprintf("Waiting for %s (up to %ds)", opts.label, timeoutSec))

	deadline := time.Now().Add(opts.timeout)
	start := time.Now()
	var lastErr error

	for time.Now().Before(deadline) {
		if err := ctx.Err(); err != nil {
			return err
		}

		err := agent.runRemoteProbe(script)
		if err == nil {
			util.PrintInfo(fmt.Sprintf("%s is ready", opts.label))
			return nil
		}
		lastErr = err

		if util.IsSSHSignalTerminated(err) {
			if ok, verifyErr := agent.verifyHostAfterSignalTerminate(ctx); verifyErr != nil {
				util.SSHVerbose(fmt.Sprintf("verify after SSH 143 during %s poll: %v", opts.label, verifyErr))
			} else if ok {
				util.PrintInfo(fmt.Sprintf("%s is ready", opts.label))
				return nil
			}
		} else {
			util.SSHVerbose(fmt.Sprintf("probe %s failed: %v", script, err))
		}

		elapsedSec := int(time.Since(start).Seconds())
		util.PrintInfo(fmt.Sprintf("Waiting for %s... (%ds / %ds)", opts.label, elapsedSec, timeoutSec))

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(opts.interval):
		}
	}

	msg := fmt.Sprintf("timed out waiting for %s after %ds", opts.label, timeoutSec)
	if lastErr != nil {
		msg = fmt.Sprintf("%s: %v (check edgelet and edgelet-containerd on the host)", msg, lastErr)
	}
	return fmt.Errorf("%s", msg)
}

func (agent *RemoteEdgelet) pollContainerdReady(ctx context.Context) error {
	return agent.pollRemoteScript(ctx, pkg.edgeletScriptProbeContainerdReady, pollOptions{
		interval: 2 * time.Second,
		timeout:  120 * time.Second,
		label:    "containerd socket",
	})
}

func (agent *RemoteEdgelet) pollEdgeletReady(ctx context.Context) error {
	return agent.pollRemoteScript(ctx, pkg.edgeletScriptProbeEdgeletReady, pollOptions{
		interval: 5 * time.Second,
		timeout:  600 * time.Second,
		label:    "edgelet daemon",
	})
}

func (agent *RemoteEdgelet) verifyHostAfterSignalTerminate(ctx context.Context) (healthy bool, err error) {
	if remoteVerifyHostHook != nil {
		return remoteVerifyHostHook(agent)
	}
	if remoteEdgeletRunHook != nil {
		return true, nil
	}

	verifyCtx, cancel := context.WithTimeout(ctx, verifyProbeTimeout)
	defer cancel()

	if err := agent.runRemoteProbeWithContext(verifyCtx, pkg.edgeletScriptProbeContainerdReady); err == nil {
		util.PrintInfo("SSH session ended (143) but host reports ready; continuing")
		return true, nil
	}

	if err := agent.runRemoteProbeWithContext(verifyCtx, pkg.edgeletScriptProbeEdgeletReady); err == nil {
		util.PrintInfo("SSH session ended (143) but host reports ready; continuing")
		return true, nil
	}

	return false, nil
}

func (agent *RemoteEdgelet) runRemoteProbeWithContext(ctx context.Context, script string) error {
	if remotePollScriptHook != nil {
		return remotePollScriptHook(agent, script)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return agent.runRemoteProbe(script)
}

func (agent *RemoteEdgelet) runStartEdgeletNoWait(ctx context.Context) error {
	env := agent.remoteBootstrapEnv()
	startEdgelet := Entrypoint{
		destPath: agent.procs.StartEdgelet.destPath,
		Args:     append([]string{"--no-wait"}, append([]string(nil), agent.procs.StartEdgelet.Args...)...),
	}
	cmdStr := wrapBootstrapCommand("sudo "+startEdgelet.getCommand(), env, true)

	Verbose("Starting edgelet on " + agent.name)
	if remoteEdgeletRunHook != nil {
		return agent.run([]command{{cmd: cmdStr, msg: "Starting edgelet on " + agent.name}})
	}

	if err := agent.ssh.Connect(); err != nil {
		return err
	}
	defer util.Log(agent.ssh.Disconnect)

	_, err := agent.ssh.RunWithOptions(cmdStr, util.SSHRunOptions{StreamOutput: true})
	if err == nil {
		return nil
	}
	if util.IsSSHSignalTerminated(err) {
		if ok, _ := agent.verifyHostAfterSignalTerminate(ctx); ok {
			util.PrintInfo("start_edgelet SSH closed (143) but engine started; continuing")
			return nil
		}
		return fmt.Errorf("start edgelet: %w", err)
	}
	return err
}

func (agent *RemoteEdgelet) postInstallRemote(ctx context.Context) error {
	env := agent.remoteBootstrapEnv()
	withEnv := func(cmd string) string {
		return wrapBootstrapCommand(cmd, env, true)
	}
	prefix := "sudo "

	if err := agent.run([]command{{
		cmd: withEnv(prefix + agent.procs.InstallInitUnits.getCommand()),
		msg: "Installing edgelet init units on " + agent.name,
	}}); err != nil {
		return err
	}

	if err := agent.runStartEdgeletNoWait(ctx); err != nil {
		return err
	}

	if agent.cfg.containerEngine() == "edgelet" {
		if err := agent.pollContainerdReady(ctx); err != nil {
			return err
		}
	}

	if err := agent.run([]command{{
		cmd: withEnv(prefix + agent.procs.ConfigureContainer.getCommand()),
		msg: "Configuring edgelet container on " + agent.name,
	}}); err != nil {
		return err
	}

	return agent.pollEdgeletReady(ctx)
}
