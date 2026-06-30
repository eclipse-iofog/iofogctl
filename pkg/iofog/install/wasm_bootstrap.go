package install

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install/wasm"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

const wasmRemoteStageSubdir = "wasm"

type wasmManifestEntry struct {
	Handler string `json:"handler"`
	Src     string `json:"src"`
	Name    string `json:"name"`
}

func (cfg EdgeletInstallConfig) wasmScope() wasm.InstallScope {
	deploy := cfg.deploymentType()
	return wasm.InstallScope{
		HostOS:          cfg.hostOS(),
		DeploymentType:  deploy,
		ContainerEngine: cfg.containerEngine(),
		HasWasm:         len(cfg.Wasm) > 0,
	}
}

func (cfg EdgeletInstallConfig) wasmPlatform() string {
	arch := strings.TrimSpace(cfg.Arch)
	if arch == "" || arch == "auto" {
		arch = "amd64"
	}
	return cfg.hostOS() + "/" + arch
}

func wasmHandlerEnvKey(handler string) string {
	return strings.ToUpper(strings.ReplaceAll(handler, "-", "_"))
}

// PrepareWasm resolves artifacts and sets bootstrap env for install_wasm_runtimes.sh.
func (cfg *EdgeletInstallConfig) PrepareWasm(ctx context.Context, namespace string, freshInstall bool) error {
	if cfg.wasmEnv != "" {
		return nil
	}
	cfg.wasmEnv = ""
	if !wasm.ShouldInstallWasm(cfg.wasmScope()) {
		if len(cfg.Wasm) > 0 {
			util.PrintInfo("Skipping WASM runtime install (scope gate)")
		}
		return nil
	}

	staged, err := wasm.ResolveWasmArtifacts(ctx, namespace, cfg.wasmPlatform(), cfg.Wasm, cfg.Airgap)
	if err != nil {
		return err
	}
	if len(staged) == 0 {
		return nil
	}

	stageDir := filepath.Join(EdgeletScriptStageDir, wasmRemoteStageSubdir)
	if err := os.MkdirAll(stageDir, util.DirPerm); err != nil {
		return err
	}

	entries := make([]wasmManifestEntry, 0, len(staged))
	for _, item := range staged {
		destName := item.CanonicalName
		destPath := filepath.Join(stageDir, destName)
		if err := copyWasmStageBinary(item.LocalPath, destPath); err != nil {
			return fmt.Errorf("stage WASM shim %s: %w", item.Handler, err)
		}
		entries = append(entries, wasmManifestEntry{
			Handler: item.Handler,
			Src:     destPath,
			Name:    destName,
		})
	}
	return cfg.applyWasmManifest(entries, staged, freshInstall, stageDir)
}

// SetWasmStaged configures bootstrap env from pre-staged binaries (local or remote airgap paths).
func (cfg *EdgeletInstallConfig) SetWasmStaged(staged []wasm.StagedBinary, freshInstall bool, remotePaths bool) error {
	if !wasm.ShouldInstallWasm(cfg.wasmScope()) {
		if len(cfg.Wasm) > 0 {
			util.PrintInfo("Skipping WASM runtime install (scope gate)")
		}
		return nil
	}
	if len(staged) == 0 {
		return nil
	}

	entries := make([]wasmManifestEntry, 0, len(staged))
	for _, item := range staged {
		src := item.LocalPath
		if remotePaths {
			src = item.RemotePath
		}
		if src == "" {
			return fmt.Errorf("WASM shim %s is missing staged path", item.Handler)
		}
		entries = append(entries, wasmManifestEntry{
			Handler: item.Handler,
			Src:     src,
			Name:    item.CanonicalName,
		})
	}
	return cfg.applyWasmManifest(entries, staged, freshInstall, "")
}

func (cfg *EdgeletInstallConfig) applyWasmManifest(entries []wasmManifestEntry, staged []wasm.StagedBinary, freshInstall bool, stageDir string) error {
	handlers := make([]string, 0, len(entries))
	anyChanged := false
	for i, entry := range entries {
		handlers = append(handlers, entry.Handler)
		if i < len(staged) && staged[i].Changed {
			anyChanged = true
		}
	}
	sort.Strings(handlers)

	parts := []string{
		"EDGELET_WASM_INSTALL=1",
		fmt.Sprintf("EDGELET_WASM_HANDLERS=%s", strings.Join(handlers, ",")),
	}
	if stageDir != "" {
		manifestPath := filepath.Join(stageDir, "manifest.json")
		manifestData, err := json.Marshal(entries)
		if err != nil {
			return err
		}
		if err := os.WriteFile(manifestPath, manifestData, util.FilePerm); err != nil {
			return err
		}
		parts = append(parts, fmt.Sprintf("EDGELET_WASM_MANIFEST=%s", manifestPath))
	}
	for _, entry := range entries {
		key := wasmHandlerEnvKey(entry.Handler)
		parts = append(parts,
			fmt.Sprintf("EDGELET_WASM_BIN_%s=%s", key, entry.Src),
			fmt.Sprintf("EDGELET_WASM_NAME_%s=%s", key, entry.Name),
		)
	}

	skipRestart := freshInstall || !wasm.NeedsEngineRestart(staged, cfg.containerEngine())
	if skipRestart {
		parts = append(parts, "EDGELET_WASM_SKIP_RESTART=1")
	} else if anyChanged {
		util.PrintInfo("WASM shims changed; engine will restart before edgelet start")
	} else {
		util.PrintInfo("WASM shims unchanged")
	}

	cfg.wasmEnv = strings.Join(parts, " ")
	return nil
}

func copyWasmStageBinary(src, dest string) error {
	in, err := util.OpenValidatedFile(src)
	if err != nil {
		return err
	}
	defer util.IgnoreClose(in)

	if err := os.MkdirAll(filepath.Dir(dest), util.DirPerm); err != nil {
		return err
	}

	out, err := util.CreateUserFile(dest, util.ExecPerm)
	if err != nil {
		return err
	}
	defer util.IgnoreClose(out)

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}

func localEngineActive(cfg EdgeletInstallConfig) bool {
	engine := cfg.containerEngine()
	switch engine {
	case "edgelet":
		out, err := util.Exec("", "sh", "-c", "systemctl is-active edgelet-containerd 2>/dev/null || true")
		return err == nil && strings.TrimSpace(out.String()) == "active"
	case "docker":
		_, err := util.Exec("", "sh", "-c", "docker ps >/dev/null 2>&1")
		return err == nil
	default:
		return false
	}
}
