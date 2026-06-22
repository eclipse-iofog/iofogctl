package install

import (
	"strings"
	"testing"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func TestRemoteEdgeletBootstrapUsesMockedSSH(t *testing.T) {
	util.SetEdgeletReleaseBaseForTest("https://example.com/download")
	util.SetEdgeletBinaryVersionForTest("v1.0.0-rc.3")
	t.Cleanup(func() {
		util.ResetEdgeletReleaseBaseForTest()
		util.ResetEdgeletBinaryVersionForTest()
	})

	var ran []string
	remoteEdgeletRunHook = func(agent *RemoteEdgelet, cmds []command) error {
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
	if len(ran) < 8 {
		t.Fatalf("expected bootstrap commands, got %d: %v", len(ran), ran)
	}

	joined := strings.Join(ran, " ")
	if !strings.Contains(joined, "install.sh") {
		t.Fatalf("expected install.sh in bootstrap, got: %v", ran)
	}
	if !strings.Contains(joined, "wait_edgelet_ready.sh") {
		t.Fatalf("expected wait_edgelet_ready.sh in bootstrap, got: %v", ran)
	}
	if !strings.Contains(joined, "/etc/edgelet") {
		t.Fatalf("expected runtime config materialization, got: %v", ran)
	}
}

func TestRemoteEdgeletConfigureUsesMockedSSH(t *testing.T) {
	var ran []string
	remoteEdgeletRunHook = func(agent *RemoteEdgelet, cmds []command) error {
		for _, cmd := range cmds {
			ran = append(ran, cmd.cmd)
		}
		return nil
	}
	t.Cleanup(func() { remoteEdgeletRunHook = nil })

	agent := &RemoteEdgelet{
		defaultAgent: defaultAgent{name: "edge-node", uuid: "agent-uuid"},
		cfg:          EdgeletInstallConfig{DeploymentType: "native"},
	}

	getProvisionKeyHook = func(*defaultAgent, string, IofogUser) (string, string, error) {
		return "provision-key", "base64-ca", nil
	}
	t.Cleanup(func() { getProvisionKeyHook = nil })

	if _, err := agent.Configure("https://controller.example.com", IofogUser{}); err != nil {
		t.Fatalf("Configure: %v", err)
	}
	if len(ran) != 3 {
		t.Fatalf("expected 3 edgelet commands, got %d: %v", len(ran), ran)
	}
	for _, want := range []string{"edgelet config --a", "edgelet config cert", "edgelet provision"} {
		found := false
		for _, cmd := range ran {
			if strings.Contains(cmd, want) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing %q in commands: %v", want, ran)
		}
	}
}
