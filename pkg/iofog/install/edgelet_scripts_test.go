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
		"--upgrade",
		"installed_embed_hash",
		"record_restart_data_plane_decision",
	} {
		if !strings.Contains(content, needle) {
			t.Fatalf("install.sh missing OTA parity marker %q", needle)
		}
	}
	if !strings.Contains(content, "Skipping daemon start (--skip-start); use start_edgelet.sh") {
		t.Fatalf("install.sh missing --skip-start handling on upgrade path")
	}

	embedLib, err := loadEdgeletScript("lib/embed.sh")
	if err != nil {
		t.Fatalf("load lib/embed.sh: %v", err)
	}
	for _, needle := range []string{
		"installed_embed_hash",
		"binary_embed_hash",
		"should_restart_data_plane",
		"start_edgelet_containerd_unit",
		"write_restart_data_plane_marker",
		"consume_restart_data_plane_marker",
	} {
		if !strings.Contains(embedLib, needle) {
			t.Fatalf("lib/embed.sh missing %q", needle)
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
		"consume_containerd_restarted_marker",
		"consume_restart_data_plane_marker",
		"Thin OTA (embed hash unchanged)",
		"drain may take up to 120s",
		"restart --no-block edgelet-containerd",
		"wait_edgelet_containerd_ready",
		"waiting for edgelet-containerd",
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
	if !strings.Contains(start, "redeploy_native_linux") || !strings.Contains(start, "start_embedded_systemd") {
		t.Fatalf("start_edgelet.sh missing redeploy stop/start edgelet parity")
	}
	if !strings.Contains(start, "consume_restart_data_plane_marker") {
		t.Fatalf("start_edgelet.sh missing embed-hash redeploy handling")
	}
	if !strings.Contains(start, "wait_edgelet_containerd_socket") {
		t.Fatalf("start_edgelet.sh missing embedded systemd start parity")
	}

	waitReady, err := loadEdgeletScript(pkg.edgeletScriptWaitEdgeletReady)
	if err != nil {
		t.Fatalf("load wait_edgelet_ready.sh: %v", err)
	}
	if !strings.Contains(waitReady, "wait_for_expected_daemon_version") {
		t.Fatalf("wait_edgelet_ready.sh missing daemon.version receipt check")
	}
	for _, needle := range []string{
		"LOCAL_API_STARTING",
		"Local API is starting",
		"runtime.engineReady",
		"edgelet_api_socket_ready",
	} {
		if !strings.Contains(waitReady, needle) {
			t.Fatalf("wait_edgelet_ready.sh missing v1.0.1 readiness marker %q", needle)
		}
	}

	wasmInstall, err := loadEdgeletScript(pkg.edgeletScriptInstallWasmRuntimes)
	if err != nil {
		t.Fatalf("load install_wasm_runtimes.sh: %v", err)
	}
	if !strings.Contains(wasmInstall, "restart_edgelet_containerd_service") {
		t.Fatalf("install_wasm_runtimes.sh must use restart_edgelet_containerd_service")
	}
	if strings.Contains(wasmInstall, "systemctl restart edgelet-containerd") {
		t.Fatalf("install_wasm_runtimes.sh must not call systemctl restart inline")
	}
	if !strings.Contains(wasmInstall, "EDGELET_WASM_DEFER_ENGINE_RESTART") {
		t.Fatalf("install_wasm_runtimes.sh must honor deferred engine restart from potctl")
	}
}

func TestEdgeletRedeployInstallFlags(t *testing.T) {
	cfg := EdgeletInstallConfig{
		HostOS:          "linux",
		ContainerEngine: "edgelet",
		Version:         "v1.0.1",
	}
	cfg.SetRedeployState(true, "v1.0.0")
	flags, err := cfg.nativeInstallFlags()
	if err != nil {
		t.Fatalf("nativeInstallFlags: %v", err)
	}
	joined := strings.Join(flags, " ")
	if !strings.Contains(joined, "--upgrade") {
		t.Fatalf("expected --upgrade for version bump redeploy, got %v", flags)
	}

	cfg.SetRedeployState(true, "v1.0.1")
	flags, err = cfg.nativeInstallFlags()
	if err != nil {
		t.Fatalf("nativeInstallFlags: %v", err)
	}
	joined = strings.Join(flags, " ")
	if strings.Contains(joined, "--upgrade") {
		t.Fatalf("did not expect --upgrade for same-version redeploy, got %v", flags)
	}

	cfg.SetRedeployState(true, "")
	flags, err = cfg.nativeInstallFlags()
	if err != nil {
		t.Fatalf("nativeInstallFlags: %v", err)
	}
	joined = strings.Join(flags, " ")
	if strings.Contains(joined, "--upgrade") {
		t.Fatalf("did not expect --upgrade when installed version is unknown, got %v", flags)
	}
}

func TestEdgeletRedeployBootstrapEnv(t *testing.T) {
	cfg := EdgeletInstallConfig{
		HostOS:          "linux",
		ContainerEngine: "edgelet",
		Version:         "v1.0.1",
	}
	cfg.SetRedeployState(true, "v1.0.1")
	env := cfg.bootstrapEnv(true)
	if !strings.Contains(env, "EDGELET_SERVICE_ACTION=restart") {
		t.Fatalf("expected redeploy restart action in bootstrap env, got %q", env)
	}
	if strings.Contains(env, "EDGELET_SERVICE_ACTION=upgrade") {
		t.Fatalf("did not expect upgrade action for same-version redeploy, got %q", env)
	}

	cfg.SetRedeployState(true, "v1.0.0")
	env = cfg.bootstrapEnv(true)
	if !strings.Contains(env, "EDGELET_SERVICE_ACTION=upgrade") {
		t.Fatalf("expected upgrade action for version bump redeploy, got %q", env)
	}

	cfg.SetRedeployState(false, "")
	env = cfg.bootstrapEnv(true)
	if strings.Contains(env, "EDGELET_SERVICE_ACTION=restart") {
		t.Fatalf("did not expect restart action on fresh install, got %q", env)
	}
}
