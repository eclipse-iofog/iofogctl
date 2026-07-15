package install

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install/wasm"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

// LocalEdgelet installs edgelet on the local host using layered scripts.
type LocalEdgelet struct {
	defaultAgent
	dir           string
	shareDir      string
	procs         EdgeletProcedures
	cfg           EdgeletInstallConfig
	customInstall bool
}

func NewLocalEdgelet(name, agentUUID string, cfg EdgeletInstallConfig) (*LocalEdgelet, error) {
	stageDir := EdgeletScriptStageDir
	procs, err := newDefaultEdgeletProcedures(stageDir, cfg)
	if err != nil {
		return nil, err
	}
	return &LocalEdgelet{
		defaultAgent: defaultAgent{name: name, uuid: agentUUID},
		dir:          stageDir,
		shareDir:     EdgeletShareDir(cfg.hostOS()),
		procs:        procs,
		cfg:          cfg,
	}, nil
}

func (agent *LocalEdgelet) CustomizeProcedures(dir string, procs *EdgeletProcedures) error {
	dir, err := util.FormatPath(dir)
	if err != nil {
		return err
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		procs.scriptNames = append(procs.scriptNames, file.Name())
		content, err := util.ReadFileUnderRoot(dir, file.Name())
		if err != nil {
			return err
		}
		procs.scriptContents = append(procs.scriptContents, string(content))
	}

	procs.scriptNames = append(procs.scriptNames, pkg.edgeletScriptPrereq)
	prereqContent, err := loadEdgeletScript(pkg.edgeletScriptPrereq)
	if err != nil {
		return err
	}
	procs.scriptContents = append(procs.scriptContents, prereqContent)

	if procs.Deps.Name == "" {
		procs.Deps = agent.procs.Deps
		for _, script := range []string{
			pkg.edgeletScriptInstallDeps,
			pkg.edgeletScriptConfigureContainerEngine,
		} {
			procs.scriptNames = append(procs.scriptNames, script)
			scriptContent, err := loadEdgeletScript(script)
			if err != nil {
				return err
			}
			procs.scriptContents = append(procs.scriptContents, scriptContent)
		}
	}
	if procs.Install.Name == "" {
		procs.Install = agent.procs.Install
		for _, script := range []string{
			pkg.edgeletScriptInstall,
			pkg.edgeletScriptInstallWasmRuntimes,
			pkg.edgeletScriptInstallContainer,
			pkg.edgeletScriptInstallInitUnits,
			pkg.edgeletScriptStartEdgelet,
			pkg.edgeletScriptConfigureContainerEdgelet,
			pkg.edgeletScriptWaitEdgeletReady,
			pkg.edgeletScriptBundled,
		} {
			procs.scriptNames = append(procs.scriptNames, script)
			scriptContent, err := loadEdgeletScript(script)
			if err != nil {
				return err
			}
			procs.scriptContents = append(procs.scriptContents, scriptContent)
		}
		for _, script := range pkg.edgeletLibScripts {
			procs.scriptNames = append(procs.scriptNames, script)
			scriptContent, err := loadEdgeletScript(script)
			if err != nil {
				return err
			}
			procs.scriptContents = append(procs.scriptContents, scriptContent)
		}
	} else {
		agent.customInstall = true
	}
	if procs.InstallInitUnits.Name == "" {
		procs.InstallInitUnits = agent.procs.InstallInitUnits
	}
	if procs.InstallWasmRuntimes.Name == "" {
		procs.InstallWasmRuntimes = agent.procs.InstallWasmRuntimes
	}
	if procs.StartEdgelet.Name == "" {
		procs.StartEdgelet = agent.procs.StartEdgelet
	}
	if procs.ConfigureContainer.Name == "" {
		procs.ConfigureContainer = agent.procs.ConfigureContainer
	}
	if procs.WaitEdgeletReady.Name == "" {
		procs.WaitEdgeletReady = agent.procs.WaitEdgeletReady
	}
	if procs.Bundled.Name == "" {
		procs.Bundled = agent.procs.Bundled
	}
	if procs.Uninstall.Name == "" {
		procs.Uninstall = agent.procs.Uninstall
		procs.scriptNames = append(procs.scriptNames, pkg.edgeletScriptUninstall)
		scriptContent, err := loadEdgeletScript(pkg.edgeletScriptUninstall)
		if err != nil {
			return err
		}
		procs.scriptContents = append(procs.scriptContents, scriptContent)
	}

	agent.bindProcedurePaths(procs, agent.dir)
	agent.procs = *procs
	return nil
}

func (agent *LocalEdgelet) bindProcedurePaths(procs *EdgeletProcedures, stageDir string) {
	procs.check.destPath = util.JoinAgentPath(stageDir, pkg.edgeletScriptPrereq)
	procs.DetectInit.destPath = util.JoinAgentPath(stageDir, pkg.edgeletScriptDetectInit)
	procs.Deps.destPath = util.JoinAgentPath(stageDir, procs.Deps.Name)
	procs.refreshInstallEntry(stageDir, agent.cfg)
	procs.InstallWasmRuntimes.destPath = util.JoinAgentPath(stageDir, pkg.edgeletScriptInstallWasmRuntimes)
	procs.InstallInitUnits.destPath = util.JoinAgentPath(stageDir, pkg.edgeletScriptInstallInitUnits)
	procs.InstallContainer.destPath = util.JoinAgentPath(stageDir, pkg.edgeletScriptInstallContainer)
	procs.StartEdgelet.destPath = util.JoinAgentPath(stageDir, pkg.edgeletScriptStartEdgelet)
	procs.ConfigureContainer.destPath = util.JoinAgentPath(stageDir, pkg.edgeletScriptConfigureContainerEdgelet)
	procs.WaitEdgeletReady.destPath = util.JoinAgentPath(stageDir, pkg.edgeletScriptWaitEdgeletReady)
	procs.Bundled.destPath = util.JoinAgentPath(stageDir, pkg.edgeletScriptBundled)
	procs.Uninstall.destPath = util.JoinAgentPath(stageDir, procs.Uninstall.Name)
}

func (agent *LocalEdgelet) SetVersion(version string) error {
	if version == "" || agent.customInstall {
		return nil
	}
	agent.cfg.Version = version
	return agent.procs.setInstallArgs(agent.cfg)
}

func (agent *LocalEdgelet) SetContainerImage(image string) error {
	if image == "" || agent.customInstall {
		return nil
	}
	agent.cfg.ContainerImage = image
	return agent.procs.setInstallArgs(agent.cfg)
}

func (agent *LocalEdgelet) SetAirgap(binPath string) error {
	agent.cfg.Airgap = true
	agent.cfg.BinPath = binPath
	return agent.procs.setInstallArgs(agent.cfg)
}

func (agent *LocalEdgelet) PrepareWasm(ctx context.Context, namespace string) error {
	if err := refreshLocalRedeployState(&agent.cfg, &agent.procs); err != nil {
		return err
	}
	freshInstall := !agent.cfg.engineActive
	return agent.cfg.PrepareWasm(ctx, namespace, freshInstall)
}

func (agent *LocalEdgelet) SetWasmStaged(staged []wasm.StagedBinary) error {
	if err := refreshLocalRedeployState(&agent.cfg, &agent.procs); err != nil {
		return err
	}
	freshInstall := !agent.cfg.engineActive
	return agent.cfg.SetWasmStaged(staged, freshInstall, false)
}

func (agent *LocalEdgelet) Bootstrap() error {
	if err := refreshLocalRedeployState(&agent.cfg, &agent.procs); err != nil {
		return err
	}
	if err := agent.materializeScripts(); err != nil {
		return err
	}
	useSudo := needsLocalSudo(agent.cfg)
	for _, cmd := range agent.procs.preInstallCommands(agent.name, agent.cfg, useSudo) {
		Verbose(cmd.msg)
		if err := agent.runShell(cmd.cmd); err != nil {
			return err
		}
	}
	if err := agent.restartDeferredWasmEngine(); err != nil {
		return err
	}
	if err := agent.finalizeDeferredWasmEngine(); err != nil {
		return err
	}
	if ShouldMaterializeEdgeletRuntime(agent.cfg) {
		if err := MaterializeEdgeletRuntime(agent.cfg.hostOS(), agent.cfg.Runtime); err != nil {
			return err
		}
	}
	for _, cmd := range agent.procs.postInstallCommandsBeforeBundled(agent.name, agent.cfg, useSudo) {
		Verbose(cmd.msg)
		if err := agent.runShell(cmd.cmd); err != nil {
			return err
		}
	}
	if err := agent.DeployWasmRuntimeClasses(); err != nil {
		return err
	}
	bundled := agent.procs.postInstallBundledCommand(agent.name, agent.cfg, useSudo)
	Verbose(bundled.msg)
	if err := agent.runShell(bundled.cmd); err != nil {
		return err
	}
	return nil
}

func (agent *LocalEdgelet) Deprovision() error {
	cmd := agent.edgeletCommand("edgelet deprovision", "Deprovisioning edgelet on "+agent.name)
	if err := agent.runShell(cmd); err != nil && !isEdgeletNotProvisionedError(err) {
		return err
	}
	return nil
}

func (agent *LocalEdgelet) Prune() error {
	return agent.runShell(agent.edgeletCommand("edgelet system prune", "Pruning edgelet on "+agent.name))
}

func (agent *LocalEdgelet) Uninstall(removeData bool) error {
	if err := agent.materializeScripts(); err != nil {
		return err
	}
	agent.procs.setUninstallArgs(agent.cfg, removeData)
	cmd := agent.procs.Uninstall.getCommand()
	if needsLocalSudo(agent.cfg) {
		cmd = "sudo env PATH=$PATH:/usr/local/bin:/usr/bin:/sbin " + cmd
	}
	msg := "Removing edgelet from " + agent.name
	Verbose(msg)
	return agent.runShell(agent.cfg.bootstrapEnv(true) + " " + cmd)
}

func (agent *LocalEdgelet) edgeletCommand(cmd, msg string) string {
	prefix := ""
	if needsLocalSudo(agent.cfg) {
		prefix = "sudo env PATH=$PATH:/usr/local/bin:/usr/bin:/sbin "
	}
	Verbose(msg)
	return agent.cfg.bootstrapEnv(true) + " " + prefix + cmd
}

func (agent *LocalEdgelet) Configure(controllerEndpoint string, user IofogUser, sdkOpt client.Options) (string, error) {
	key, caCert, err := agent.getProvisionKey(controllerEndpoint, user, sdkOpt)
	if err != nil {
		return "", err
	}
	cmds, err := EdgeletProvisionCommands(controllerEndpoint, key, caCert, needsLocalSudo(agent.cfg))
	if err != nil {
		return "", err
	}
	for _, cmd := range cmds {
		Verbose(cmd.msg)
		if err := agent.runShell(agent.edgeletCommand(cmd.cmd, cmd.msg)); err != nil {
			return "", err
		}
	}
	return agent.uuid, nil
}

func needsLocalSudo(cfg EdgeletInstallConfig) bool {
	return cfg.native()
}

func (agent *LocalEdgelet) materializeScripts() error {
	if err := os.MkdirAll(agent.dir, util.DirPerm); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(agent.dir, "lib"), util.DirPerm); err != nil {
		return err
	}
	for idx, script := range agent.procs.scriptNames {
		path := filepath.Join(agent.dir, script)
		if err := os.MkdirAll(filepath.Dir(path), util.DirPerm); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(agent.procs.scriptContents[idx]), util.ExecPerm); err != nil { // #nosec G306 -- executable install scripts
			return err
		}
	}
	return nil
}

func (agent *LocalEdgelet) runShell(command string) error {
	if strings.TrimSpace(command) == "" {
		return fmt.Errorf("empty command")
	}
	// Match remote SSH: full command string runs under a shell so shellJoinArgs quotes work.
	_, err := util.Exec("", "sh", "-c", command)
	return err
}
