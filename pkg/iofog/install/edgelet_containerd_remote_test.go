package install

import (
	"strings"
	"testing"
)

func TestRestartDeferredWasmEngine_MarksAfterPoll(t *testing.T) {
	var ran []string
	remoteEdgeletRunHook = func(_ *RemoteEdgelet, cmds []command) error {
		for _, cmd := range cmds {
			ran = append(ran, cmd.cmd)
		}
		return nil
	}
	t.Cleanup(func() { remoteEdgeletRunHook = nil })

	remotePollScriptHook = func(_ *RemoteEdgelet, script string) error {
		if script != pkg.edgeletScriptProbeContainerdReady {
			t.Fatalf("unexpected probe script %q", script)
		}
		return nil
	}
	t.Cleanup(func() { remotePollScriptHook = nil })

	agent := &RemoteEdgelet{
		defaultAgent: defaultAgent{name: "edge-node"},
		dir:          EdgeletScriptStageDir,
		cfg: EdgeletInstallConfig{
			ContainerEngine:        "edgelet",
			engineActive:           true,
			wasmDeferEngineRestart: true,
		},
	}

	if err := agent.restartDeferredWasmEngine(); err != nil {
		t.Fatalf("restartDeferredWasmEngine: %v", err)
	}

	joined := strings.Join(ran, " ")
	if !strings.Contains(joined, pkg.edgeletScriptRestartContainerd) {
		t.Fatalf("expected restart script in commands, got %v", ran)
	}
	if !strings.Contains(joined, ".containerd-restarted") {
		t.Fatalf("expected mark containerd restarted after poll, got %v", ran)
	}
}
