package install

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func TestLocalEdgeletRunShellHonorsShellQuotes(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "echoarg.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho \"$1\"\n"), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}

	// runShell delegates to sh -c; quoted args must not reach argv with literal quotes.
	out, err := util.Exec("", "sh", "-c", script+" 'docker'")
	if err != nil {
		t.Fatalf("Exec: %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "docker" {
		t.Fatalf("got %q, want docker", got)
	}
}

func TestLocalEdgeletPreInstallDepsCommandQuoted(t *testing.T) {
	cfg := EdgeletInstallConfig{
		HostOS:          "darwin",
		ContainerEngine: "docker",
		DeploymentType:  "native",
	}
	procs, err := newDefaultEdgeletProcedures(EdgeletScriptStageDir, cfg)
	if err != nil {
		t.Fatalf("newDefaultEdgeletProcedures: %v", err)
	}
	pre := procs.preInstallCommands("edge-2", cfg, false)
	if len(pre) < 3 {
		t.Fatalf("expected pre-install commands, got %d", len(pre))
	}
	depsCmd := pre[2].cmd
	if !strings.Contains(depsCmd, "install_deps.sh 'docker' 'native'") {
		t.Fatalf("expected shell-quoted deps args, got %q", depsCmd)
	}
}
