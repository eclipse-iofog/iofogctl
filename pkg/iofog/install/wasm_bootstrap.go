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
// stageKey isolates operator-local staging per agent so concurrent deploys do not collide.
func (cfg *EdgeletInstallConfig) PrepareWasm(ctx context.Context, namespace, stageKey string, freshInstall bool) error {
	if cfg.wasmEnv != "" {
		return nil
	}
	cfg.wasmEnv = ""
	cfg.wasmLocalStageDir = ""
	cfg.wasmDeferEngineRestart = false
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

	stageDir := WasmLocalStageDir(stageKey)
	if err := os.MkdirAll(stageDir, util.DirPerm); err != nil {
		return err
	}

	keep := make(map[string]struct{}, len(staged))
	entries := make([]wasmManifestEntry, 0, len(staged))
	for _, item := range staged {
		destName := item.CanonicalName
		keep[destName] = struct{}{}
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
	if err := pruneWasmStageDir(stageDir, keep); err != nil {
		return err
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
	cfg.wasmDeferEngineRestart = false

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
		cfg.wasmLocalStageDir = stageDir
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

	plan := PlanEngineRestart(*cfg, hostStateFromWasm(freshInstall, anyChanged))
	skipRestart := freshInstall || !wasm.NeedsEngineRestart(staged, cfg.containerEngine()) || !anyChanged
	if skipRestart {
		parts = append(parts, "EDGELET_WASM_SKIP_RESTART=1")
	} else if plan.Deferred && plan.Needed {
		cfg.wasmDeferEngineRestart = true
		parts = append(parts, "EDGELET_WASM_DEFER_ENGINE_RESTART=1")
		util.PrintInfo("WASM shims changed; edgelet-containerd restart may take up to 2 minutes — do not interrupt")
	} else {
		util.PrintInfo("WASM shims unchanged")
	}

	cfg.wasmEnv = strings.Join(parts, " ")
	return nil
}

func pruneWasmStageDir(stageDir string, keep map[string]struct{}) error {
	entries, err := os.ReadDir(stageDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name == "manifest.json" {
			continue
		}
		if _, ok := keep[name]; ok {
			continue
		}
		if err := os.Remove(filepath.Join(stageDir, name)); err != nil {
			return err
		}
	}
	return nil
}

func readWasmManifest(path string) ([]wasmManifestEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var entries []wasmManifestEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
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
		ready, _ := containerdReadyProbe(EdgeletScriptStageDir)
		return ready
	case "docker":
		_, err := util.Exec("", "sh", "-c", "docker ps >/dev/null 2>&1")
		return err == nil
	default:
		return false
	}
}

func readInstalledEdgeletVersion() string {
	out, err := util.Exec("", "sh", "-c", installedEdgeletVersionShell())
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out.String())
}

func refreshLocalRedeployState(cfg *EdgeletInstallConfig, procs *EdgeletProcedures) error {
	active := localEngineActive(*cfg)
	version := ""
	if active {
		version = readInstalledEdgeletVersion()
	}
	cfg.SetRedeployState(active, version)
	if procs != nil {
		return procs.setInstallArgs(*cfg)
	}
	return nil
}
