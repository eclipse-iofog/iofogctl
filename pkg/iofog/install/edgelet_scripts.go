package install

import (
	"fmt"
	"os"
	"strings"

	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install/wasm"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

// EdgeletInstallConfig drives layered install script arguments.
type EdgeletInstallConfig struct {
	HostOS                 string
	Version                string
	Arch                   string
	ContainerEngine        string
	DeploymentType         string
	ContainerImage         string
	TimeZone               string
	BinPath                string
	Airgap                 bool
	Wasm                   map[string]wasm.Pack
	Runtime                *EdgeletRuntimeSpec
	wasmEnv                string
	engineActive           bool
	installedVersion       string
	wasmDeferEngineRestart bool
}

const edgeletInstallReceiptPath = "/var/backups/edgelet/install-receipt"

// installedEdgeletVersionShell reads installed_version from receipt (sudo) with edgelet --version fallback.
func installedEdgeletVersionShell() string {
	return `v=$(sudo grep '^installed_version=' ` + edgeletInstallReceiptPath + ` 2>/dev/null | head -1 | sed 's/^installed_version=//'); if [ -z "$v" ]; then v=$(edgelet --version 2>/dev/null | awk -F': ' '$1 == "daemon.version" {print $2; exit}' | tr -d '[:space:]'); fi; printf '%s' "$v"`
}

// SetRedeployState records whether edgelet is already running on the host and the installed version from receipt.
func (cfg *EdgeletInstallConfig) SetRedeployState(engineActive bool, installedVersion string) {
	cfg.engineActive = engineActive
	cfg.installedVersion = strings.TrimSpace(installedVersion)
}

func (cfg EdgeletInstallConfig) needsUpgradeInstall() bool {
	if !cfg.engineActive {
		return false
	}
	target := cfg.version()
	if cfg.installedVersion == "" {
		// Receipt unreadable or missing — do not stop edgelet for a spurious upgrade.
		return false
	}
	return cfg.installedVersion != target
}

func (cfg EdgeletInstallConfig) native() bool {
	return cfg.DeploymentType == "" || cfg.DeploymentType == "native"
}

func (cfg EdgeletInstallConfig) containerEngine() string {
	if cfg.ContainerEngine == "" {
		return "edgelet"
	}
	return cfg.ContainerEngine
}

func (cfg EdgeletInstallConfig) deploymentType() string {
	if cfg.DeploymentType == "" {
		return "native"
	}
	return cfg.DeploymentType
}

func (cfg EdgeletInstallConfig) version() string {
	if cfg.Version != "" {
		return cfg.Version
	}
	return util.GetEdgeletBinaryVersion()
}

func (cfg EdgeletInstallConfig) installModeEnv() string {
	if cfg.native() {
		return "native"
	}
	return "container"
}

func (cfg EdgeletInstallConfig) containerImage() string {
	if cfg.ContainerImage != "" {
		return cfg.ContainerImage
	}
	return util.GetEdgeletImage()
}

func (cfg EdgeletInstallConfig) timeZone() string {
	if cfg.TimeZone != "" {
		return cfg.TimeZone
	}
	return "UTC"
}

func (cfg EdgeletInstallConfig) depsArgs() []string {
	return []string{cfg.containerEngine(), cfg.deploymentType()}
}

func (cfg EdgeletInstallConfig) hostOS() string {
	if cfg.HostOS != "" {
		return cfg.HostOS
	}
	return "linux"
}

func (cfg EdgeletInstallConfig) shareDir() string {
	return EdgeletShareDir(cfg.hostOS())
}

func (cfg EdgeletInstallConfig) containerEngineURL() string {
	engine := cfg.containerEngine()
	if cfg.Runtime != nil {
		return resolveContainerEngineURL(engine, &cfg.Runtime.Agent)
	}
	return resolveContainerEngineURL(engine, nil)
}

func (cfg EdgeletInstallConfig) bootstrapEnv(localInstall bool) string {
	parts := []string{
		fmt.Sprintf("EDGELET_INSTALL_MODE=%s", cfg.installModeEnv()),
		fmt.Sprintf("CONTAINER_ENGINE=%s", cfg.containerEngine()),
		fmt.Sprintf("DEPLOYMENT_TYPE=%s", cfg.deploymentType()),
		fmt.Sprintf("EDGELET_VERSION=%s", cfg.version()),
		fmt.Sprintf("EDGELET_CONTAINER_IMAGE=%s", cfg.containerImage()),
		fmt.Sprintf("EDGELET_TZ=%s", cfg.timeZone()),
		fmt.Sprintf("EDGELET_GITHUB_REPO=%s", util.GetEdgeletGitHubRepo()),
		fmt.Sprintf("EDGELET_CONTAINER_ENGINE_URL=%s", cfg.containerEngineURL()),
	}
	if localInstall {
		parts = append(parts, "LOCAL_INSTALL=1")
	}
	if localInstall && IsDesktopContainerDeploy(cfg) {
		parts = append(parts, fmt.Sprintf("EDGELET_SCRIPT_STAGE_DIR=%s", EdgeletScriptStageDir))
		pathEnv := os.Getenv("PATH")
		if pathEnv == "" {
			pathEnv = "/usr/bin:/bin"
		}
		parts = append(parts, fmt.Sprintf("PATH=%s/bin:%s", EdgeletScriptStageDir, pathEnv))
	}
	if IsDesktopContainerDeploy(cfg) {
		if cmd, err := cfg.bootstrapConfigCommand(); err == nil && cmd != "" {
			parts = append(parts, fmt.Sprintf("EDGELET_BOOTSTRAP_CONFIG_CMD=%s", shellQuoteArg(cmd)))
		}
	}
	if cfg.wasmEnv != "" {
		parts = append(parts, cfg.wasmEnv)
	}
	if cfg.engineActive {
		if cfg.needsUpgradeInstall() {
			parts = append(parts, "EDGELET_SERVICE_ACTION=upgrade")
		} else {
			parts = append(parts, "EDGELET_SERVICE_ACTION=restart")
		}
	}
	return strings.Join(parts, " ")
}

func (cfg EdgeletInstallConfig) nativeInstallFlags() ([]string, error) {
	flags := []string{
		fmt.Sprintf("--version=%s", cfg.version()),
		fmt.Sprintf("--container-engine=%s", cfg.containerEngine()),
		"--skip-config",
		"--skip-start",
	}
	if cfg.needsUpgradeInstall() {
		flags = append([]string{"--upgrade"}, flags...)
	}
	if cfg.Arch != "" && cfg.Arch != "auto" {
		flags = append(flags, fmt.Sprintf("--arch=%s", cfg.Arch))
	}
	if cfg.Airgap {
		flags = append(flags, "--airgap")
		if cfg.BinPath != "" {
			flags = append(flags, fmt.Sprintf("--bin-path=%s", cfg.BinPath))
		}
	}
	return flags, nil
}

func (cfg EdgeletInstallConfig) containerInstallFlags() []string {
	return []string{
		fmt.Sprintf("--image=%s", cfg.containerImage()),
		fmt.Sprintf("--engine=%s", cfg.containerEngine()),
		fmt.Sprintf("--tz=%s", cfg.timeZone()),
	}
}

func (cfg EdgeletInstallConfig) installFlags() ([]string, error) {
	if cfg.native() {
		return cfg.nativeInstallFlags()
	}
	return cfg.containerInstallFlags(), nil
}

func (cfg EdgeletInstallConfig) uninstallArgs(removeData bool) []string {
	if !removeData {
		return nil
	}
	return []string{"--remove-data"}
}

// EdgeletProcedures extends AgentProcedures with edgelet bootstrap layers.
type EdgeletProcedures struct {
	AgentProcedures
	DetectInit          Entrypoint
	InstallContainer    Entrypoint
	InstallWasmRuntimes Entrypoint
	InstallInitUnits    Entrypoint
	StartEdgelet        Entrypoint
	ConfigureContainer  Entrypoint
	WaitEdgeletReady    Entrypoint
	Bundled             Entrypoint
}

func edgeletScriptNames() []string {
	names := []string{
		pkg.edgeletScriptPrereq,
		pkg.edgeletScriptDetectInit,
		pkg.edgeletScriptInstallDeps,
		pkg.edgeletScriptConfigureContainerEngine,
		pkg.edgeletScriptInstall,
		pkg.edgeletScriptInstallWasmRuntimes,
		pkg.edgeletScriptInstallContainer,
		pkg.edgeletScriptInstallInitUnits,
		pkg.edgeletScriptStartEdgelet,
		pkg.edgeletScriptConfigureContainerEdgelet,
		pkg.edgeletScriptWaitEdgeletReady,
		pkg.edgeletScriptBundled,
		pkg.edgeletScriptUninstall,
	}
	return append(names, pkg.edgeletLibScripts...)
}

func addEdgeletAssetPrefix(file string) string {
	return "edgelet/scripts/" + file
}

func loadEdgeletScript(name string) (string, error) {
	return util.GetStaticFile(addEdgeletAssetPrefix(name))
}

func loadDefaultEdgeletScripts() ([]string, []string, error) {
	names := edgeletScriptNames()
	contents := make([]string, 0, len(names))
	for _, name := range names {
		content, err := loadEdgeletScript(name)
		if err != nil {
			return nil, nil, err
		}
		contents = append(contents, content)
	}
	return names, contents, nil
}

func newDefaultEdgeletProcedures(dir string, cfg EdgeletInstallConfig) (EdgeletProcedures, error) {
	installFlags, err := cfg.installFlags()
	if err != nil {
		return EdgeletProcedures{}, err
	}

	installEntry := Entrypoint{
		Name:     pkg.edgeletScriptInstall,
		destPath: util.JoinAgentPath(dir, pkg.edgeletScriptInstall),
		Args:     installFlags,
	}
	if !cfg.native() {
		installEntry = Entrypoint{
			Name:     pkg.edgeletScriptInstallContainer,
			destPath: util.JoinAgentPath(dir, pkg.edgeletScriptInstallContainer),
			Args:     installFlags,
		}
	}

	procs := EdgeletProcedures{
		AgentProcedures: AgentProcedures{
			check: Entrypoint{
				Name:     pkg.edgeletScriptPrereq,
				destPath: util.JoinAgentPath(dir, pkg.edgeletScriptPrereq),
			},
			Deps: Entrypoint{
				Name:     pkg.edgeletScriptInstallDeps,
				destPath: util.JoinAgentPath(dir, pkg.edgeletScriptInstallDeps),
				Args:     cfg.depsArgs(),
			},
			Install: installEntry,
			Uninstall: Entrypoint{
				Name:     pkg.edgeletScriptUninstall,
				destPath: util.JoinAgentPath(dir, pkg.edgeletScriptUninstall),
			},
		},
		DetectInit: Entrypoint{
			Name:     pkg.edgeletScriptDetectInit,
			destPath: util.JoinAgentPath(dir, pkg.edgeletScriptDetectInit),
		},
		InstallContainer: Entrypoint{
			Name:     pkg.edgeletScriptInstallContainer,
			destPath: util.JoinAgentPath(dir, pkg.edgeletScriptInstallContainer),
			Args:     cfg.containerInstallFlags(),
		},
		InstallWasmRuntimes: Entrypoint{
			Name:     pkg.edgeletScriptInstallWasmRuntimes,
			destPath: util.JoinAgentPath(dir, pkg.edgeletScriptInstallWasmRuntimes),
		},
		InstallInitUnits: Entrypoint{
			Name:     pkg.edgeletScriptInstallInitUnits,
			destPath: util.JoinAgentPath(dir, pkg.edgeletScriptInstallInitUnits),
		},
		StartEdgelet: Entrypoint{
			Name:     pkg.edgeletScriptStartEdgelet,
			destPath: util.JoinAgentPath(dir, pkg.edgeletScriptStartEdgelet),
		},
		ConfigureContainer: Entrypoint{
			Name:     pkg.edgeletScriptConfigureContainerEdgelet,
			destPath: util.JoinAgentPath(dir, pkg.edgeletScriptConfigureContainerEdgelet),
		},
		WaitEdgeletReady: Entrypoint{
			Name:     pkg.edgeletScriptWaitEdgeletReady,
			destPath: util.JoinAgentPath(dir, pkg.edgeletScriptWaitEdgeletReady),
		},
		Bundled: Entrypoint{
			Name:     pkg.edgeletScriptBundled,
			destPath: util.JoinAgentPath(dir, pkg.edgeletScriptBundled),
		},
	}

	names, contents, err := loadDefaultEdgeletScripts()
	if err != nil {
		return EdgeletProcedures{}, err
	}
	procs.scriptNames = names
	procs.scriptContents = contents
	return procs, nil
}

func (procs *EdgeletProcedures) setInstallArgs(cfg EdgeletInstallConfig) error {
	flags, err := cfg.installFlags()
	if err != nil {
		return err
	}
	procs.refreshInstallEntry(stageDirFromEntrypoint(procs.Install.destPath), cfg)
	procs.Install.Args = flags
	procs.InstallContainer.Args = cfg.containerInstallFlags()
	return nil
}

func stageDirFromEntrypoint(destPath string) string {
	if destPath == "" {
		return EdgeletScriptStageDir
	}
	if idx := strings.LastIndex(destPath, "/"); idx > 0 {
		return destPath[:idx]
	}
	return EdgeletScriptStageDir
}

func (procs *EdgeletProcedures) refreshInstallEntry(stageDir string, cfg EdgeletInstallConfig) {
	if cfg.native() {
		procs.Install.Name = pkg.edgeletScriptInstall
		procs.Install.destPath = util.JoinAgentPath(stageDir, pkg.edgeletScriptInstall)
		return
	}
	procs.Install.Name = pkg.edgeletScriptInstallContainer
	procs.Install.destPath = util.JoinAgentPath(stageDir, pkg.edgeletScriptInstallContainer)
}

func (procs *EdgeletProcedures) setUninstallArgs(cfg EdgeletInstallConfig, removeData bool) {
	procs.Uninstall.Args = cfg.uninstallArgs(removeData)
}

func (procs *EdgeletProcedures) setDepsArgs(cfg EdgeletInstallConfig) {
	procs.Deps.Args = cfg.depsArgs()
}

// wrapBootstrapCommand prefixes bootstrap env vars. When useSudo is true, env is passed via
// "sudo env ..." so macOS/Linux sudo does not strip CONTAINER_ENGINE and related vars.
func wrapBootstrapCommand(cmd, env string, useSudo bool) string {
	if env == "" {
		return cmd
	}
	if useSudo && strings.HasPrefix(cmd, "sudo ") {
		return "sudo env " + env + " " + strings.TrimPrefix(cmd, "sudo ")
	}
	return env + " " + cmd
}

func (procs *EdgeletProcedures) preInstallCommands(name string, cfg EdgeletInstallConfig, useSudo bool) []command {
	prefix := ""
	if useSudo {
		prefix = "sudo "
	}
	env := cfg.bootstrapEnv(!useSudo)
	withEnv := func(cmd string) string {
		return wrapBootstrapCommand(cmd, env, useSudo)
	}
	cmds := []command{
		{cmd: withEnv(procs.check.getCommand()), msg: "Checking prerequisites on " + name},
		{cmd: withEnv(procs.DetectInit.getCommand()), msg: "Detecting OS/init on " + name},
		{cmd: withEnv(procs.Deps.getCommand()), msg: "Installing dependencies on " + name},
		{cmd: withEnv(prefix + procs.Install.getCommand()), msg: "Installing edgelet on " + name},
	}
	if wasm.ShouldInstallWasm(cfg.wasmScope()) {
		cmds = append(cmds, command{
			cmd: withEnv(prefix + procs.InstallWasmRuntimes.getCommand()),
			msg: "Installing WASM runtimes on " + name,
		})
	}
	return cmds
}

func (procs *EdgeletProcedures) postInstallCommandsBeforeBundled(name string, cfg EdgeletInstallConfig, useSudo bool) []command {
	prefix := ""
	if useSudo {
		prefix = "sudo "
	}
	env := cfg.bootstrapEnv(!useSudo)
	withEnv := func(cmd string) string {
		return wrapBootstrapCommand(cmd, env, useSudo)
	}
	return []command{
		{cmd: withEnv(prefix + procs.InstallInitUnits.getCommand()), msg: "Installing edgelet init units on " + name},
		{cmd: withEnv(prefix + procs.StartEdgelet.getCommand()), msg: "Starting edgelet on " + name},
		{cmd: withEnv(prefix + procs.ConfigureContainer.getCommand()), msg: "Configuring edgelet container on " + name},
		{cmd: withEnv(prefix + procs.WaitEdgeletReady.getCommand()), msg: "Waiting for edgelet on " + name},
	}
}

func (procs *EdgeletProcedures) postInstallBundledCommand(name string, cfg EdgeletInstallConfig, useSudo bool) command {
	prefix := ""
	if useSudo {
		prefix = "sudo "
	}
	env := cfg.bootstrapEnv(!useSudo)
	withEnv := func(cmd string) string {
		return wrapBootstrapCommand(cmd, env, useSudo)
	}
	return command{
		cmd: withEnv(prefix + procs.Bundled.getCommand()),
		msg: "Publishing edgelet scripts on " + name,
	}
}

func (procs *EdgeletProcedures) postInstallCommands(name string, cfg EdgeletInstallConfig, useSudo bool) []command {
	cmds := procs.postInstallCommandsBeforeBundled(name, cfg, useSudo)
	return append(cmds, procs.postInstallBundledCommand(name, cfg, useSudo))
}

func isEdgeletNotProvisionedError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "not provisioned")
}

func isEdgeletControlPlaneRemovedError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "not found") ||
		strings.Contains(msg, "no control plane") ||
		strings.Contains(msg, "already removed")
}
