package install

import (
	"strings"
	"testing"

	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install/wasm"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func TestEdgeletScriptLayerOrder(t *testing.T) {
	names := edgeletScriptNames()
	wantAfterInstall := -1
	wantWasm := -1
	wantInitUnits := -1
	for i, name := range names {
		switch name {
		case pkg.edgeletScriptInstall:
			wantAfterInstall = i
		case pkg.edgeletScriptInstallWasmRuntimes:
			wantWasm = i
		case pkg.edgeletScriptInstallInitUnits:
			wantInitUnits = i
		}
	}
	if wantAfterInstall < 0 || wantWasm < 0 || wantInitUnits < 0 {
		t.Fatalf("missing expected scripts in %v", names)
	}
	if wantAfterInstall >= wantWasm || wantWasm >= wantInitUnits {
		t.Fatalf("expected install -> wasm -> init_units order, got indices install=%d wasm=%d init=%d in %v",
			wantAfterInstall, wantWasm, wantInitUnits, names)
	}
}

func TestPreInstallCommandsIncludeWasmWhenConfigured(t *testing.T) {
	util.SetEdgeletReleaseBaseForTest("https://example.com/download")
	util.SetEdgeletBinaryVersionForTest("v1.0.0-rc.6")
	t.Cleanup(func() {
		util.ResetEdgeletReleaseBaseForTest()
		util.ResetEdgeletBinaryVersionForTest()
	})

	cfg := EdgeletInstallConfig{
		HostOS:          "linux",
		Arch:            "amd64",
		ContainerEngine: "edgelet",
		DeploymentType:  "native",
		Wasm: map[string]wasm.Pack{
			"spin": {URL: "https://example.invalid/spin.tar.gz"},
		},
	}
	procs, err := newDefaultEdgeletProcedures(EdgeletScriptStageDir, cfg)
	if err != nil {
		t.Fatalf("newDefaultEdgeletProcedures: %v", err)
	}

	pre := procs.preInstallCommands("edge-node", cfg, true)
	if len(pre) != 5 {
		t.Fatalf("expected 5 pre-install commands with wasm, got %d", len(pre))
	}
	if !strings.Contains(pre[4].cmd, pkg.edgeletScriptInstallWasmRuntimes) {
		t.Fatalf("expected wasm install command, got %q", pre[4].cmd)
	}

	cfg.Wasm = nil
	pre = procs.preInstallCommands("edge-node", cfg, true)
	if len(pre) != 4 {
		t.Fatalf("expected 4 pre-install commands without wasm, got %d", len(pre))
	}
}

func TestInstallScriptOTAParity(t *testing.T) {
	content, err := loadEdgeletScript(pkg.edgeletScriptInstall)
	if err != nil {
		t.Fatalf("load install.sh: %v", err)
	}
	for _, needle := range []string{
		"stop_edgelet_service",
		"EDGELET_SERVICE_ACTION",
		"edgelet_cli_version",
	} {
		if !strings.Contains(content, needle) {
			t.Fatalf("install.sh missing OTA parity marker %q", needle)
		}
	}

	service, err := loadEdgeletScript("lib/service.sh")
	if err != nil {
		t.Fatalf("load lib/service.sh: %v", err)
	}
	for _, needle := range []string{
		"stop_edgelet_service",
		"restart_edgelet_services",
		"restart_edgelet_containerd_service",
	} {
		if !strings.Contains(service, needle) {
			t.Fatalf("lib/service.sh missing %q", needle)
		}
	}
	if strings.Contains(service, `prefix="${_field}.version:"`) {
		t.Fatalf("lib/service.sh uses broken awk prefix (trailing colon with -F': ')")
	}
	if !strings.Contains(service, `key="${_field}.version"`) {
		t.Fatalf("lib/service.sh missing fixed edgelet_version_field awk key")
	}

	start, err := loadEdgeletScript(pkg.edgeletScriptStartEdgelet)
	if err != nil {
		t.Fatalf("load start_edgelet.sh: %v", err)
	}
	if !strings.Contains(start, "EDGELET_SERVICE_ACTION") || !strings.Contains(start, "restart_native_linux") {
		t.Fatalf("start_edgelet.sh missing OTA restart handling")
	}

	waitReady, err := loadEdgeletScript(pkg.edgeletScriptWaitEdgeletReady)
	if err != nil {
		t.Fatalf("load wait_edgelet_ready.sh: %v", err)
	}
	if !strings.Contains(waitReady, "wait_for_expected_daemon_version") {
		t.Fatalf("wait_edgelet_ready.sh missing daemon.version receipt check")
	}
}
