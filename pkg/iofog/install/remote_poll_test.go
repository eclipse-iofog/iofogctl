package install

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func TestPollRemoteScript_SuccessThirdTry(t *testing.T) {
	attempts := 0
	remotePollScriptHook = func(_ *RemoteEdgelet, script string) error {
		attempts++
		if script != pkg.edgeletScriptProbeContainerdReady {
			t.Fatalf("unexpected script %q", script)
		}
		if attempts < 3 {
			return errors.New("probe not ready")
		}
		return nil
	}
	t.Cleanup(func() { remotePollScriptHook = nil })

	agent := &RemoteEdgelet{dir: EdgeletScriptStageDir}
	err := agent.pollRemoteScript(context.Background(), pkg.edgeletScriptProbeContainerdReady, pollOptions{
		interval: 10 * time.Millisecond,
		timeout:  500 * time.Millisecond,
		label:    "containerd socket",
	})
	if err != nil {
		t.Fatalf("pollRemoteScript: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 probe attempts, got %d", attempts)
	}
}

func TestPollRemoteScript_143ThenSuccess(t *testing.T) {
	attempts := 0
	remotePollScriptHook = func(_ *RemoteEdgelet, script string) error {
		attempts++
		if script != pkg.edgeletScriptProbeEdgeletReady {
			t.Fatalf("unexpected script %q", script)
		}
		if attempts == 1 {
			return errors.New("Error during SSH Session\nProcess exited with status 143")
		}
		return nil
	}
	t.Cleanup(func() { remotePollScriptHook = nil })

	remoteVerifyHostHook = func(*RemoteEdgelet) (bool, error) {
		return false, nil
	}
	t.Cleanup(func() { remoteVerifyHostHook = nil })

	agent := &RemoteEdgelet{dir: EdgeletScriptStageDir}
	err := agent.pollRemoteScript(context.Background(), pkg.edgeletScriptProbeEdgeletReady, pollOptions{
		interval: 10 * time.Millisecond,
		timeout:  500 * time.Millisecond,
		label:    "edgelet daemon",
	})
	if err != nil {
		t.Fatalf("pollRemoteScript: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("expected 2 probe attempts, got %d", attempts)
	}
}

func TestPollRemoteScript_Timeout(t *testing.T) {
	remotePollScriptHook = func(*RemoteEdgelet, string) error {
		return errors.New("still not ready")
	}
	t.Cleanup(func() { remotePollScriptHook = nil })

	agent := &RemoteEdgelet{dir: EdgeletScriptStageDir}
	err := agent.pollRemoteScript(context.Background(), pkg.edgeletScriptProbeContainerdReady, pollOptions{
		interval: 20 * time.Millisecond,
		timeout:  50 * time.Millisecond,
		label:    "containerd socket",
	})
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "timed out waiting for containerd socket") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPostInstallRemote_NoWaitEdgeletReadyScript(t *testing.T) {
	util.SetEdgeletReleaseBaseForTest("https://example.com/download")
	util.SetEdgeletBinaryVersionForTest("v1.0.0-rc.8")
	t.Cleanup(func() {
		util.ResetEdgeletReleaseBaseForTest()
		util.ResetEdgeletBinaryVersionForTest()
	})

	var ran []string
	remoteEdgeletRunHook = func(_ *RemoteEdgelet, cmds []command) error {
		for _, cmd := range cmds {
			ran = append(ran, cmd.cmd)
		}
		return nil
	}
	t.Cleanup(func() { remoteEdgeletRunHook = nil })
	remoteEdgeletInstallFileHook = func(*RemoteEdgelet, string, []byte, string) error { return nil }
	t.Cleanup(func() { remoteEdgeletInstallFileHook = nil })

	cfg := EdgeletInstallConfig{
		HostOS:          "linux",
		Arch:            "amd64",
		ContainerEngine: "edgelet",
		DeploymentType:  "native",
	}
	procs, err := newDefaultEdgeletProcedures(EdgeletScriptStageDir, cfg)
	if err != nil {
		t.Fatalf("newDefaultEdgeletProcedures: %v", err)
	}
	agent := &RemoteEdgelet{
		defaultAgent: defaultAgent{name: "edge-node"},
		dir:          EdgeletScriptStageDir,
		procs:        procs,
		cfg:          cfg,
	}

	if err := agent.Bootstrap(); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}

	joined := strings.Join(ran, " ")
	if strings.Contains(joined, pkg.edgeletScriptWaitEdgeletReady) {
		t.Fatalf("remote bootstrap must not invoke wait_edgelet_ready.sh, got: %v", ran)
	}
	if !strings.Contains(joined, pkg.edgeletScriptStartEdgelet) {
		t.Fatalf("expected start_edgelet.sh in bootstrap, got: %v", ran)
	}
	if !strings.Contains(joined, "--no-wait") {
		t.Fatalf("expected --no-wait in start_edgelet command, got: %v", ran)
	}
	if !strings.Contains(joined, pkg.edgeletScriptInstallInitUnits) {
		t.Fatalf("expected install_init_units.sh in bootstrap, got: %v", ran)
	}
	if !strings.Contains(joined, pkg.edgeletScriptConfigureContainerEdgelet) {
		t.Fatalf("expected configure_container_edgelet.sh in bootstrap, got: %v", ran)
	}
}

func TestPostInstallRemote_UsesProbeScripts(t *testing.T) {
	util.SetEdgeletReleaseBaseForTest("https://example.com/download")
	util.SetEdgeletBinaryVersionForTest("v1.0.0-rc.8")
	t.Cleanup(func() {
		util.ResetEdgeletReleaseBaseForTest()
		util.ResetEdgeletBinaryVersionForTest()
	})

	var ran []string
	remoteEdgeletRunHook = func(_ *RemoteEdgelet, cmds []command) error {
		for _, cmd := range cmds {
			ran = append(ran, cmd.cmd)
		}
		return nil
	}
	t.Cleanup(func() { remoteEdgeletRunHook = nil })

	var probes []string
	remotePollScriptHook = func(_ *RemoteEdgelet, script string) error {
		probes = append(probes, script)
		return nil
	}
	t.Cleanup(func() { remotePollScriptHook = nil })

	cfg := EdgeletInstallConfig{
		HostOS:          "linux",
		Arch:            "amd64",
		ContainerEngine: "edgelet",
		DeploymentType:  "native",
	}
	procs, err := newDefaultEdgeletProcedures(EdgeletScriptStageDir, cfg)
	if err != nil {
		t.Fatalf("newDefaultEdgeletProcedures: %v", err)
	}
	agent := &RemoteEdgelet{
		defaultAgent: defaultAgent{name: "edge-node"},
		dir:          EdgeletScriptStageDir,
		procs:        procs,
		cfg:          cfg,
	}

	if err := agent.postInstallRemote(context.Background()); err != nil {
		t.Fatalf("postInstallRemote: %v", err)
	}

	joined := strings.Join(ran, " ")
	if strings.Contains(joined, pkg.edgeletScriptWaitEdgeletReady) {
		t.Fatalf("postInstallRemote must not invoke wait_edgelet_ready.sh")
	}
	if !strings.Contains(joined, "--no-wait") {
		t.Fatalf("postInstallRemote must start edgelet with --no-wait, got: %v", ran)
	}

	probeSet := make(map[string]bool)
	for _, script := range probes {
		probeSet[script] = true
	}
	if !probeSet[pkg.edgeletScriptProbeContainerdReady] {
		t.Fatalf("postInstallRemote must poll containerd via probe script, got probes: %v", probes)
	}
	if !probeSet[pkg.edgeletScriptProbeEdgeletReady] {
		t.Fatalf("postInstallRemote must poll edgelet via probe script, got probes: %v", probes)
	}
}

func TestVerifyHostAfterSignalTerminate_ContainerdReady(t *testing.T) {
	calls := 0
	remotePollScriptHook = func(_ *RemoteEdgelet, script string) error {
		calls++
		if script == pkg.edgeletScriptProbeContainerdReady {
			return nil
		}
		return errors.New("edgelet not ready")
	}
	t.Cleanup(func() { remotePollScriptHook = nil })

	agent := &RemoteEdgelet{dir: EdgeletScriptStageDir}
	ok, err := agent.verifyHostAfterSignalTerminate(context.Background())
	if err != nil {
		t.Fatalf("verifyHostAfterSignalTerminate: %v", err)
	}
	if !ok {
		t.Fatal("expected healthy host")
	}
	if calls != 1 {
		t.Fatalf("expected single containerd probe, got %d calls", calls)
	}
}
