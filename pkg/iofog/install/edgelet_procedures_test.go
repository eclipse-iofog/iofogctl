package install

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func TestDefaultEdgeletProceduresScripts(t *testing.T) {
	util.SetEdgeletReleaseBaseForTest("https://github.com/Datasance/edgelet/releases/download")
	util.SetEdgeletBinaryVersionForTest("v1.0.0-rc.8")
	t.Cleanup(func() {
		util.ResetEdgeletReleaseBaseForTest()
		util.ResetEdgeletBinaryVersionForTest()
	})

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

	names := edgeletScriptNames()
	if len(procs.scriptNames) != len(names) {
		t.Fatalf("expected %d embedded scripts, got %d", len(names), len(procs.scriptNames))
	}
	if len(procs.scriptContents) != len(procs.scriptNames) {
		t.Fatalf("script content count mismatch")
	}
	if procs.Deps.Args[0] != "edgelet" {
		t.Fatalf("deps args = %v, want edgelet engine", procs.Deps.Args)
	}
	joined := strings.Join(procs.Install.Args, " ")
	if !strings.Contains(joined, "--version=v1.0.0-rc.8") {
		t.Fatalf("expected --version flag in install args, got %v", procs.Install.Args)
	}
	if !strings.Contains(joined, "--skip-start") {
		t.Fatalf("expected --skip-start in install args, got %v", procs.Install.Args)
	}
}

func TestEdgeletInstallFlagsContainer(t *testing.T) {
	cfg := EdgeletInstallConfig{
		DeploymentType:  "container",
		ContainerEngine: "docker",
		ContainerImage:  "ghcr.io/example/edgelet:1.2.3",
		TimeZone:        "Europe/Istanbul",
	}
	flags, err := cfg.installFlags()
	if err != nil {
		t.Fatalf("installFlags: %v", err)
	}
	joined := strings.Join(flags, " ")
	if !strings.Contains(joined, "--image=ghcr.io/example/edgelet:1.2.3") {
		t.Fatalf("unexpected container install flags: %v", flags)
	}
	if !strings.Contains(joined, "--engine=docker") {
		t.Fatalf("unexpected container install flags: %v", flags)
	}
}

func TestEdgeletBootstrapEnvContainerEngineURL(t *testing.T) {
	cfg := EdgeletInstallConfig{
		ContainerEngine: "podman",
		DeploymentType:  "container",
	}
	env := cfg.bootstrapEnv(true)
	if !strings.Contains(env, "EDGELET_CONTAINER_ENGINE_URL=unix:///run/podman/podman.sock") {
		t.Fatalf("bootstrap env missing podman socket URL: %q", env)
	}
}

func TestCustomizeEdgeletProceduresPartial(t *testing.T) {
	dir := t.TempDir()
	srcDir := "../../../assets/edgelet/scripts"
	if err := copyDir(srcDir, dir); err != nil {
		t.Fatalf("copyDir: %v", err)
	}
	if err := os.Remove(path.Join(dir, pkg.edgeletScriptPrereq)); err != nil {
		t.Fatalf("remove prereq: %v", err)
	}
	if err := os.Remove(path.Join(dir, pkg.edgeletScriptInstall)); err != nil {
		t.Fatalf("remove install script: %v", err)
	}

	agent := &RemoteEdgelet{dir: EdgeletScriptStageDir}
	procs := EdgeletProcedures{
		AgentProcedures: AgentProcedures{
			Deps: Entrypoint{Name: pkg.edgeletScriptInstallDeps},
			Uninstall: Entrypoint{
				Name: pkg.edgeletScriptUninstall,
			},
		},
	}
	if err := agent.CustomizeProcedures(dir, &procs); err != nil {
		t.Fatalf("CustomizeProcedures: %v", err)
	}
	if agent.customInstall {
		t.Fatalf("expected default install script to be embedded")
	}
	if len(procs.scriptNames) < len(edgeletScriptNames()) {
		t.Fatalf("expected embedded defaults to be appended, got %v", procs.scriptNames)
	}
}

func TestInstallDepsSkipMatrix(t *testing.T) {
	tests := []struct {
		engine string
		deploy string
		skip   bool
	}{
		{"edgelet", "native", true},
		{"edgelet", "container", true},
		{"docker", "native", false},
		{"podman", "container", false},
	}
	for _, tt := range tests {
		cfg := EdgeletInstallConfig{ContainerEngine: tt.engine, DeploymentType: tt.deploy}
		if got := util.ShouldSkipInstallDeps(cfg.containerEngine(), cfg.deploymentType()); got != tt.skip {
			t.Fatalf("engine=%q deploy=%q skip=%v want %v", tt.engine, tt.deploy, got, tt.skip)
		}
	}
}

func TestBootstrapCommandGeneration(t *testing.T) {
	util.SetEdgeletReleaseBaseForTest("https://example.com/download")
	util.SetEdgeletBinaryVersionForTest("v1.0.0-rc.8")

	cfg := EdgeletInstallConfig{HostOS: "linux", Arch: "amd64", ContainerEngine: "docker"}
	procs, err := newDefaultEdgeletProcedures(EdgeletScriptStageDir, cfg)
	if err != nil {
		t.Fatalf("newDefaultEdgeletProcedures: %v", err)
	}
	pre := procs.preInstallCommands("edge-node", cfg, true)
	if len(pre) != 4 {
		t.Fatalf("expected 4 pre-install commands without wasm, got %d", len(pre))
	}
	if !strings.Contains(pre[3].cmd, "install.sh") {
		t.Fatalf("expected install.sh command, got %q", pre[3].cmd)
	}
	post := procs.postInstallCommands("edge-node", cfg, true)
	if len(post) != 5 {
		t.Fatalf("expected 5 post-install commands, got %d", len(post))
	}
	if !strings.Contains(pre[0].cmd, "CONTAINER_ENGINE=docker") {
		t.Fatalf("expected bootstrap env injection, got %q", pre[0].cmd)
	}
	if !strings.Contains(pre[3].cmd, "sudo env ") || !strings.Contains(pre[3].cmd, "CONTAINER_ENGINE=docker") {
		t.Fatalf("expected sudo env bootstrap for install, got %q", pre[3].cmd)
	}
	if !strings.Contains(post[0].cmd, "sudo env ") || !strings.Contains(post[0].cmd, "CONTAINER_ENGINE=docker") {
		t.Fatalf("expected sudo env bootstrap for post-install, got %q", post[0].cmd)
	}
}

func TestEdgeletUninstallArgs(t *testing.T) {
	cfg := EdgeletInstallConfig{}
	procs, err := newDefaultEdgeletProcedures(EdgeletScriptStageDir, cfg)
	if err != nil {
		t.Fatalf("newDefaultEdgeletProcedures: %v", err)
	}
	procs.setUninstallArgs(cfg, true)
	if len(procs.Uninstall.Args) != 1 || procs.Uninstall.Args[0] != "--remove-data" {
		t.Fatalf("unexpected uninstall args with remove-data: %v", procs.Uninstall.Args)
	}
	procs.setUninstallArgs(cfg, false)
	if len(procs.Uninstall.Args) != 0 {
		t.Fatalf("expected no uninstall args without remove-data, got %v", procs.Uninstall.Args)
	}
}

func TestEdgeletShareDir(t *testing.T) {
	if got := EdgeletShareDir("linux"); got != "/usr/share/edgelet" {
		t.Fatalf("linux share dir = %q", got)
	}
	if got := EdgeletShareDir("darwin"); got != "/usr/local/share/edgelet" {
		t.Fatalf("darwin share dir = %q", got)
	}
	if got := EdgeletShareDir("windows"); got != `%ProgramData%\Edgelet\scripts` {
		t.Fatalf("windows share dir = %q", got)
	}
}
