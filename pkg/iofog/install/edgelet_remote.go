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

const remoteEdgeletManifestDir = "/tmp"

// remoteEdgeletRunHook is set by tests to mock SSH command execution.
var remoteEdgeletRunHook func(agent *RemoteEdgelet, cmds []command) error

// remoteEdgeletInstallFileHook is set by tests to mock remote config/cert writes.
var remoteEdgeletInstallFileHook func(agent *RemoteEdgelet, destPath string, content []byte, perm string) error

// RemoteEdgelet installs edgelet on a remote host over SSH using layered scripts.
type RemoteEdgelet struct {
	defaultAgent
	ssh           *util.SecureShellClient
	dir           string
	shareDir      string
	procs         EdgeletProcedures
	cfg           EdgeletInstallConfig
	customInstall bool
}

func NewRemoteEdgelet(user, host string, port int, privKeyFilename, agentName, agentUUID string, cfg EdgeletInstallConfig) (*RemoteEdgelet, error) {
	ssh, err := util.NewSecureShellClient(user, host, privKeyFilename)
	if err != nil {
		return nil, err
	}
	ssh.SetPort(port)

	stageDir := EdgeletScriptStageDir
	procs, err := newDefaultEdgeletProcedures(stageDir, cfg)
	if err != nil {
		return nil, err
	}

	return &RemoteEdgelet{
		defaultAgent: defaultAgent{name: agentName, uuid: agentUUID},
		ssh:          ssh,
		dir:          stageDir,
		shareDir:     EdgeletShareDir(cfg.hostOS()),
		procs:        procs,
		cfg:          cfg,
	}, nil
}

func (agent *RemoteEdgelet) CustomizeProcedures(dir string, procs *EdgeletProcedures) error {
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

func (agent *RemoteEdgelet) bindProcedurePaths(procs *EdgeletProcedures, stageDir string) {
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

func (agent *RemoteEdgelet) SetVersion(version string) error {
	if version == "" || agent.customInstall {
		return nil
	}
	agent.cfg.Version = version
	return agent.procs.setInstallArgs(agent.cfg)
}

func (agent *RemoteEdgelet) SetContainerImage(image string) error {
	if image == "" || agent.customInstall {
		return nil
	}
	agent.cfg.ContainerImage = image
	return agent.procs.setInstallArgs(agent.cfg)
}

func (agent *RemoteEdgelet) SetAirgap(binPath string) error {
	agent.cfg.Airgap = true
	agent.cfg.BinPath = binPath
	return agent.procs.setInstallArgs(agent.cfg)
}

func (agent *RemoteEdgelet) detectAndSetHostOS() error {
	if remoteEdgeletRunHook != nil {
		return nil
	}
	if err := agent.ssh.Connect(); err != nil {
		return err
	}
	defer util.Log(agent.ssh.Disconnect)

	out, err := agent.ssh.Run("uname -s")
	if err != nil {
		return fmt.Errorf("detect remote host OS: %w", err)
	}
	raw := strings.TrimSpace(out.String())
	osName, err := util.NormalizeEdgeletOS(raw)
	if err != nil {
		return fmt.Errorf("normalize remote host OS %q: %w", raw, err)
	}
	agent.cfg.HostOS = osName
	agent.shareDir = EdgeletShareDir(osName)
	agent.bindProcedurePaths(&agent.procs, agent.dir)
	return agent.procs.setInstallArgs(agent.cfg)
}

func (agent *RemoteEdgelet) PrepareWasm(ctx context.Context, namespace string) error {
	if agent.cfg.wasmEnv != "" {
		return nil
	}
	if err := agent.detectAndSetHostOS(); err != nil {
		return err
	}
	freshInstall := !agent.remoteEngineActive()
	if err := agent.cfg.PrepareWasm(ctx, namespace, freshInstall); err != nil {
		return err
	}
	return agent.copyWasmStagingToRemote()
}

func (agent *RemoteEdgelet) SetWasmStaged(staged []wasm.StagedBinary) error {
	if err := agent.detectAndSetHostOS(); err != nil {
		return err
	}
	freshInstall := !agent.remoteEngineActive()
	return agent.cfg.SetWasmStaged(staged, freshInstall, true)
}

func (agent *RemoteEdgelet) remoteEngineActive() bool {
	if remoteEdgeletRunHook != nil {
		return false
	}
	engine := agent.cfg.containerEngine()
	var checkCmd string
	switch engine {
	case "edgelet":
		checkCmd = "systemctl is-active edgelet-containerd 2>/dev/null"
	case "docker":
		checkCmd = "docker ps >/dev/null 2>&1"
	default:
		return false
	}
	if err := agent.ssh.Connect(); err != nil {
		return false
	}
	defer util.Log(agent.ssh.Disconnect)
	out, err := agent.ssh.Run(checkCmd)
	if engine == "edgelet" {
		return err == nil && strings.TrimSpace(out.String()) == "active"
	}
	return err == nil
}

func (agent *RemoteEdgelet) copyWasmStagingToRemote() error {
	if agent.cfg.wasmEnv == "" {
		return nil
	}
	if remoteEdgeletRunHook != nil {
		return nil
	}

	localDir := filepath.Join(EdgeletScriptStageDir, wasmRemoteStageSubdir)
	entries, err := os.ReadDir(localDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	remoteDir := util.JoinAgentPath(agent.dir, wasmRemoteStageSubdir)
	if err := agent.run([]command{{
		cmd: fmt.Sprintf("sudo mkdir -p %s && sudo chmod 755 %s", remoteDir, remoteDir),
		msg: "Creating remote WASM staging directory on " + agent.name,
	}}); err != nil {
		return err
	}

	if err := agent.ssh.Connect(); err != nil {
		return err
	}
	defer util.Log(agent.ssh.Disconnect)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		localPath := filepath.Join(localDir, entry.Name())
		content, err := util.ReadValidatedFile(localPath)
		if err != nil {
			return err
		}
		tmpName := "wasm-" + entry.Name() + ".upload"
		reader := strings.NewReader(string(content))
		if err := agent.ssh.CopyTo(reader, remoteEdgeletManifestDir, tmpName, "0755", int64(len(content))); err != nil {
			return err
		}
		destPath := util.JoinAgentPath(remoteDir, entry.Name())
		installCmd := fmt.Sprintf("sudo install -m 755 %s/%s %s", remoteEdgeletManifestDir, tmpName, destPath)
		if _, err := agent.ssh.Run(installCmd); err != nil {
			return err
		}
		if _, err := agent.ssh.Run(fmt.Sprintf("rm -f %s/%s", remoteEdgeletManifestDir, tmpName)); err != nil {
			return err
		}
	}
	return nil
}

func (agent *RemoteEdgelet) Bootstrap() error {
	if err := agent.detectAndSetHostOS(); err != nil {
		return err
	}
	if err := agent.copyInstallScripts(); err != nil {
		return err
	}
	if err := agent.run(agent.procs.preInstallCommands(agent.name, agent.cfg, true)); err != nil {
		return err
	}
	if err := agent.materializeRuntimeConfig(); err != nil {
		return err
	}
	if err := agent.run(agent.procs.postInstallCommandsBeforeBundled(agent.name, agent.cfg, true)); err != nil {
		return err
	}
	if err := agent.DeployWasmRuntimeClasses(); err != nil {
		return err
	}
	return agent.run([]command{agent.procs.postInstallBundledCommand(agent.name, agent.cfg, true)})
}

func (agent *RemoteEdgelet) Configure(controllerEndpoint string, user IofogUser, sdkOpt client.Options) (string, error) {
	key, caCert, err := agent.getProvisionKey(controllerEndpoint, user, sdkOpt)
	if err != nil {
		return "", err
	}
	cmds, err := EdgeletProvisionCommands(controllerEndpoint, key, caCert, true)
	if err != nil {
		return "", err
	}
	if err := agent.run(cmds); err != nil {
		return "", err
	}
	return agent.uuid, nil
}

func (agent *RemoteEdgelet) materializeRuntimeConfig() error {
	if !ShouldMaterializeEdgeletRuntime(agent.cfg) {
		return nil
	}
	paths := EdgeletPaths(agent.cfg.hostOS())
	if err := agent.run([]command{{
		cmd: fmt.Sprintf("sudo mkdir -p %s && sudo chmod 755 %s", paths.ConfigDir, paths.ConfigDir),
		msg: "Creating edgelet config directory on " + agent.name,
	}}); err != nil {
		return err
	}

	configYAML, err := buildRemoteEdgeletConfigYAML(agent)
	if err != nil {
		return err
	}
	if err := agent.installRemoteFileIfMissing(paths.ConfigFile, configYAML, "640"); err != nil {
		return fmt.Errorf("write edgelet config: %w", err)
	}

	sampleCA, err := loadEdgeletSampleCA()
	if err != nil {
		return err
	}
	if err := agent.installRemoteFileIfMissing(paths.CertFile, sampleCA, "644"); err != nil {
		return fmt.Errorf("write edgelet sample CA: %w", err)
	}
	return nil
}

func buildRemoteEdgeletConfigYAML(agent *RemoteEdgelet) ([]byte, error) {
	spec := agent.cfg.Runtime
	var agentCfg *client.AgentConfiguration
	arch := agent.cfg.Arch
	latitude := 0.0
	longitude := 0.0
	if spec != nil {
		agentCfg = &spec.Agent
		latitude = spec.Latitude
		longitude = spec.Longitude
		if spec.Arch != "" {
			arch = spec.Arch
		}
	}
	return BuildEdgeletConfigYAML(agent.cfg.hostOS(), arch, latitude, longitude, agentCfg)
}

func (agent *RemoteEdgelet) installRemoteFileIfMissing(destPath string, content []byte, perm string) error {
	if remoteEdgeletInstallFileHook != nil {
		return remoteEdgeletInstallFileHook(agent, destPath, content, perm)
	}
	if err := agent.ssh.Connect(); err != nil {
		return err
	}
	defer util.Log(agent.ssh.Disconnect)

	checkCmd := fmt.Sprintf("test -f %q", destPath)
	if _, err := agent.ssh.Run(checkCmd); err == nil {
		return nil
	}

	tmpName := filepath.Base(destPath) + ".tmp"
	reader := strings.NewReader(string(content))
	if err := agent.ssh.CopyTo(reader, "/tmp", tmpName, "0644", int64(len(content))); err != nil {
		return err
	}
	installCmd := fmt.Sprintf("sudo install -m %s /tmp/%s %s", perm, tmpName, destPath)
	if _, err := agent.ssh.Run(installCmd); err != nil {
		return err
	}
	_, err := agent.ssh.Run(fmt.Sprintf("rm -f /tmp/%s", tmpName))
	return err
}

// DeployFromFile applies an edgelet manifest (Registry or ControlPlane) on the remote host.
func (agent *RemoteEdgelet) DeployFromFile(manifestPath string) error {
	cmd := fmt.Sprintf("sudo edgelet deploy -f %s", shellQuoteArg(manifestPath))
	return agent.run([]command{{
		cmd: cmd,
		msg: "Deploying edgelet manifest from " + manifestPath,
	}})
}

// RegistryList runs edgelet registry ls on the remote host and returns stdout.
func (agent *RemoteEdgelet) RegistryList() (string, error) {
	if remoteEdgeletRunHook != nil {
		return "", nil
	}
	if err := agent.ssh.Connect(); err != nil {
		return "", err
	}
	defer util.Log(agent.ssh.Disconnect)

	out, err := agent.ssh.Run("sudo edgelet registry ls")
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

// WriteDeployManifest writes manifest bytes to a remote temp file for edgelet deploy -f.
func (agent *RemoteEdgelet) WriteDeployManifest(data []byte, prefix string) (path string, cleanup func(), err error) {
	localPath, localCleanup, err := WriteTempManifest(data, prefix, agent.cfg)
	if err != nil {
		return "", nil, err
	}
	defer localCleanup()

	remotePath := fmt.Sprintf("%s/edgelet-%s.yaml", remoteEdgeletManifestDir, prefix)
	if err := agent.copyLocalFileToRemote(localPath, remotePath); err != nil {
		return "", nil, err
	}
	cleanup = func() {
		_ = agent.run([]command{{
			cmd: fmt.Sprintf("rm -f %s", shellQuoteArg(remotePath)),
			msg: "Removing edgelet manifest on " + agent.name,
		}})
	}
	return remotePath, cleanup, nil
}

func (agent *RemoteEdgelet) copyLocalFileToRemote(localPath, remotePath string) error {
	if remoteEdgeletRunHook != nil {
		return nil
	}
	content, err := util.ReadValidatedFile(localPath)
	if err != nil {
		return err
	}
	if err := agent.ssh.Connect(); err != nil {
		return err
	}
	defer util.Log(agent.ssh.Disconnect)

	tmpName := filepath.Base(remotePath) + ".upload"
	reader := strings.NewReader(string(content))
	if err := agent.ssh.CopyTo(reader, remoteEdgeletManifestDir, tmpName, "0644", int64(len(content))); err != nil {
		return err
	}
	installCmd := fmt.Sprintf("sudo install -m 644 %s/%s %s", remoteEdgeletManifestDir, tmpName, remotePath)
	if _, err := agent.ssh.Run(installCmd); err != nil {
		return err
	}
	_, err = agent.ssh.Run(fmt.Sprintf("rm -f %s/%s", remoteEdgeletManifestDir, tmpName))
	return err
}

func (agent *RemoteEdgelet) Deprovision() error {
	cmds := []command{{
		cmd: "sudo edgelet deprovision",
		msg: "Deprovisioning edgelet on " + agent.name,
	}}
	if err := agent.run(cmds); err != nil && !isEdgeletNotProvisionedError(err) {
		return err
	}
	return nil
}

func (agent *RemoteEdgelet) DeleteControlPlane() error {
	cmds := []command{{
		cmd: "sudo edgelet controlplane delete",
		msg: "Deleting edgelet control plane on " + agent.name,
	}}
	if err := agent.run(cmds); err != nil && !isEdgeletControlPlaneRemovedError(err) {
		return err
	}
	return nil
}

func (agent *RemoteEdgelet) Prune() error {
	cmds := []command{{
		cmd: "sudo edgelet system prune",
		msg: "Pruning edgelet on " + agent.name,
	}}
	return agent.run(cmds)
}

func (agent *RemoteEdgelet) Uninstall(removeData bool) error {
	if err := agent.detectAndSetHostOS(); err != nil {
		return err
	}
	if err := agent.copyInstallScripts(); err != nil {
		return err
	}
	agent.procs.setUninstallArgs(agent.cfg, removeData)
	cmds := []command{
		{cmd: "sudo " + agent.procs.Uninstall.getCommand(), msg: "Removing edgelet from " + agent.name},
	}
	return agent.run(cmds)
}

func (agent *RemoteEdgelet) run(cmds []command) error {
	if remoteEdgeletRunHook != nil {
		return remoteEdgeletRunHook(agent, cmds)
	}
	if err := agent.ssh.Connect(); err != nil {
		return err
	}
	defer util.Log(agent.ssh.Disconnect)

	for _, cmd := range cmds {
		Verbose(cmd.msg)
		if _, err := agent.ssh.Run(cmd.cmd); err != nil {
			return err
		}
	}
	return nil
}

func (agent *RemoteEdgelet) copyInstallScripts() error {
	Verbose("Copying edgelet install scripts to " + agent.name)
	stage := agent.dir
	cmds := []command{
		{
			cmd: fmt.Sprintf("sudo mkdir -p %s %s/lib && sudo chmod -R 755 %s", stage, stage, stage),
			msg: "Creating edgelet script staging directory",
		},
	}
	if err := agent.run(cmds); err != nil {
		return err
	}
	if remoteEdgeletRunHook != nil {
		return nil
	}

	if err := agent.ssh.Connect(); err != nil {
		return err
	}
	defer util.Log(agent.ssh.Disconnect)

	for idx, script := range agent.procs.scriptNames {
		if err := agent.installRemoteScript(stage, script, agent.procs.scriptContents[idx]); err != nil {
			return err
		}
	}
	return nil
}

func edgeletScriptTmpName(relPath string) string {
	safe := strings.ReplaceAll(relPath, "/", "-")
	return "edgelet-" + safe + ".tmp"
}

func (agent *RemoteEdgelet) installRemoteScript(stageDir, relPath, content string) error {
	destPath := util.JoinAgentPath(stageDir, relPath)
	if strings.Contains(relPath, "/") {
		destDir := stageDir + "/" + filepath.Dir(relPath)
		mkdirCmd := fmt.Sprintf("sudo mkdir -p %s && sudo chmod 755 %s", destDir, destDir)
		if _, err := agent.ssh.Run(mkdirCmd); err != nil {
			return err
		}
	}

	tmpName := edgeletScriptTmpName(relPath)
	reader := strings.NewReader(content)
	if err := agent.ssh.CopyTo(reader, "/tmp", tmpName, "0644", int64(len(content))); err != nil {
		return err
	}
	installCmd := fmt.Sprintf("sudo install -m 755 /tmp/%s %s", tmpName, destPath)
	if _, err := agent.ssh.Run(installCmd); err != nil {
		return err
	}
	_, err := agent.ssh.Run(fmt.Sprintf("rm -f /tmp/%s", tmpName))
	return err
}
