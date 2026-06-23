package install

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

const (
	controllerAssetPrefixContainer = "container-controller"
	controllerAssetPrefixAirgap    = "airgap-controller"
)

type RemoteSystemImages struct {
	ARM     string `yaml:"arm,omitempty"`
	AMD64   string `yaml:"amd64,omitempty"`
	ARM64   string `yaml:"arm64,omitempty"`
	RISCV64 string `yaml:"riscv64,omitempty"`
}

type RemoteSystemMicroservices struct {
	Router RemoteSystemImages `yaml:"router,omitempty"`
	Nats   RemoteSystemImages `yaml:"nats,omitempty"`
}

type ControllerOptions struct {
	User                string
	Host                string
	Port                int
	Namespace           string
	PrivKeyFilename     string
	Version             string
	Image               string
	SystemMicroservices RemoteSystemMicroservices
	NatsEnabled         *bool // NATS enabling for remote control plane (nil = default enabled)
	Vault               *VaultConfig
	PidBaseDir          string
	EcnViewerPort       int
	EcnViewerURL        string
	LogLevel            string
	Https               *Https
	SiteCA              *SiteCertificate
	LocalCA             *SiteCertificate
	Airgap              bool
}

type Https struct {
	Enabled *bool
	CACert  string
	TLSCert string
	TLSKey  string
}

type SiteCertificate struct {
	TLSCert string
	TLSKey  string
}

type database struct {
	databaseName string
	provider     string
	host         string
	user         string
	password     string
	port         int
	ssl          *bool
	ca           *string
}

type auth struct {
	url              string
	realm            string
	ssl              string
	realmKey         string
	controllerClient string
	controllerSecret string
	viewerClient     string
}

type events struct {
	auditEnabled     *bool // nil if not configured, pointer to bool if configured
	retentionDays    int
	cleanupInterval  int
	captureIpAddress *bool // nil if not configured, pointer to bool if configured
}

type ControllerProcedures struct {
	check          Entrypoint `yaml:"-"` // Check prereqs script (runs for default and custom procedures)
	Deps           Entrypoint `yaml:"deps,omitempty"`
	SetEnv         Entrypoint `yaml:"setEnv,omitempty"`
	Install        Entrypoint `yaml:"install,omitempty"`
	Uninstall      Entrypoint `yaml:"uninstall,omitempty"`
	scriptNames    []string   `yaml:"-"` // List of all script names to be pushed to Controller
	scriptContents []string   `yaml:"-"` // List of contents of scripts to be pushed to Controller
}

type Controller struct {
	*ControllerOptions
	ssh           *util.SecureShellClient
	db            database
	auth          auth
	events        events
	ctrlDir       string
	iofogDir      string
	procs         ControllerProcedures
	customInstall bool // Flag set when custom install scripts are provided
	// svcDir   string
}

func NewController(options *ControllerOptions) (*Controller, error) {
	ssh, err := util.NewSecureShellClient(options.User, options.Host, options.PrivKeyFilename)
	if err != nil {
		return nil, err
	}
	ssh.SetPort(options.Port)
	if options.Image == "" {
		options.Image = util.GetControllerImage()
	}
	ctrlDir := pkg.controllerDir
	ctrl := &Controller{
		ControllerOptions: options,
		ssh:               ssh,
		iofogDir:          pkg.iofogDir,
		ctrlDir:           ctrlDir,
		procs: ControllerProcedures{
			check: Entrypoint{
				Name:     pkg.controllerScriptPrereq,
				destPath: fmt.Sprintf("%s/%s", ctrlDir, pkg.controllerScriptPrereq),
			},
			Deps: Entrypoint{
				Name:     pkg.controllerScriptInstallContainerEngine,
				destPath: fmt.Sprintf("%s/%s", ctrlDir, pkg.controllerScriptInstallContainerEngine),
			},
			SetEnv: Entrypoint{
				Name:     pkg.controllerScriptSetEnv,
				destPath: fmt.Sprintf("%s/%s", ctrlDir, pkg.controllerScriptSetEnv),
			},
			Install: Entrypoint{
				Name:     pkg.controllerScriptInstall,
				destPath: fmt.Sprintf("%s/%s", ctrlDir, pkg.controllerScriptInstall),
				Args: []string{
					options.Image,
					"",
					"",
				},
			},
			Uninstall: Entrypoint{
				Name:     pkg.controllerScriptUninstall,
				destPath: fmt.Sprintf("%s/%s", ctrlDir, pkg.controllerScriptUninstall),
			},
			scriptNames: []string{
				pkg.controllerScriptPrereq,
				pkg.controllerScriptInit,
				pkg.controllerScriptInstallContainerEngine,
				pkg.controllerScriptInstall,
				pkg.controllerScriptSetEnv,
				pkg.controllerScriptUninstall,
			},
		},
	}
	// Get script contents from embedded files
	for _, scriptName := range ctrl.procs.scriptNames {
		scriptContent, err := util.GetStaticFile(ctrl.addControllerAssetPrefix(scriptName))
		if err != nil {
			return nil, err
		}
		ctrl.procs.scriptContents = append(ctrl.procs.scriptContents, scriptContent)
	}
	return ctrl, nil
}

func (ctrl *Controller) SetControllerExternalDatabase(host, user, password, provider, databaseName string, port int, ssl *bool, ca *string) {

	ctrl.db = database{
		databaseName: databaseName,
		provider:     provider,
		host:         host,
		user:         user,
		password:     password,
		port:         port,
		ssl:          ssl,
		ca:           ca,
	}
}

func (ctrl *Controller) SetControllerAuth(url, realm, ssl, realmKey, controllerClient, controllerSecret, viewerClient string) {

	ctrl.auth = auth{
		url:              url,
		realm:            realm,
		ssl:              ssl,
		realmKey:         realmKey,
		controllerClient: controllerClient,
		controllerSecret: controllerSecret,
		viewerClient:     viewerClient,
	}
}

func (ctrl *Controller) SetControllerEvents(auditEnabled bool, retentionDays, cleanupInterval int, captureIpAddress bool) {
	ctrl.events = events{
		auditEnabled:     &auditEnabled,
		retentionDays:    retentionDays,
		cleanupInterval:  cleanupInterval,
		captureIpAddress: &captureIpAddress,
	}
}

func (ctrl *Controller) addControllerAssetPrefix(file string) string {
	if ctrl.Airgap {
		return fmt.Sprintf("%s/%s", controllerAssetPrefixAirgap, file)
	}
	return fmt.Sprintf("%s/%s", controllerAssetPrefixContainer, file)
}

func (ctrl *Controller) CustomizeProcedures(dir string, procs *ControllerProcedures) error {
	// Format source directory of script files
	dir, err := util.FormatPath(dir)
	if err != nil {
		return err
	}

	// Load script files into memory
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		if !file.IsDir() {
			procs.scriptNames = append(procs.scriptNames, file.Name())
			content, err := os.ReadFile(filepath.Join(dir, file.Name()))
			if err != nil {
				return err
			}
			procs.scriptContents = append(procs.scriptContents, string(content))
		}
	}

	// Add check_prereqs script and entrypoint (always required for both default and custom)
	procs.scriptNames = append(procs.scriptNames, pkg.controllerScriptPrereq)
	prereqContent, err := util.GetStaticFile(ctrl.addControllerAssetPrefix(pkg.controllerScriptPrereq))
	if err != nil {
		return err
	}
	procs.scriptContents = append(procs.scriptContents, prereqContent)
	procs.check.destPath = fmt.Sprintf("%s/%s", ctrl.ctrlDir, pkg.controllerScriptPrereq)

	// Add default entrypoints and scripts if necessary (user not provided)
	if procs.Deps.Name == "" {
		procs.Deps = ctrl.procs.Deps
		procs.scriptNames = append(procs.scriptNames, pkg.controllerScriptInstallContainerEngine)
		scriptContent, err := util.GetStaticFile(ctrl.addControllerAssetPrefix(pkg.controllerScriptInstallContainerEngine))
		if err != nil {
			return err
		}
		procs.scriptContents = append(procs.scriptContents, scriptContent)
	}
	if procs.SetEnv.Name == "" {
		procs.SetEnv = ctrl.procs.SetEnv
		procs.scriptNames = append(procs.scriptNames, pkg.controllerScriptSetEnv)
		scriptContent, err := util.GetStaticFile(ctrl.addControllerAssetPrefix(pkg.controllerScriptSetEnv))
		if err != nil {
			return err
		}
		procs.scriptContents = append(procs.scriptContents, scriptContent)
	}
	if procs.Install.Name == "" {
		procs.Install = ctrl.procs.Install
		procs.scriptNames = append(procs.scriptNames, pkg.controllerScriptInstall)
		scriptContent, err := util.GetStaticFile(ctrl.addControllerAssetPrefix(pkg.controllerScriptInstall))
		if err != nil {
			return err
		}
		procs.scriptContents = append(procs.scriptContents, scriptContent)
	} else {
		ctrl.customInstall = true
	}
	if procs.Uninstall.Name == "" {
		procs.Uninstall = ctrl.procs.Uninstall
		procs.scriptNames = append(procs.scriptNames, pkg.controllerScriptUninstall)
		scriptContent, err := util.GetStaticFile(ctrl.addControllerAssetPrefix(pkg.controllerScriptUninstall))
		if err != nil {
			return err
		}
		procs.scriptContents = append(procs.scriptContents, scriptContent)
	}

	// Set destination paths where scripts appear on Controller
	procs.Deps.destPath = fmt.Sprintf("%s/%s", ctrl.ctrlDir, procs.Deps.Name)
	procs.SetEnv.destPath = fmt.Sprintf("%s/%s", ctrl.ctrlDir, procs.SetEnv.Name)
	procs.Install.destPath = fmt.Sprintf("%s/%s", ctrl.ctrlDir, procs.Install.Name)
	procs.Uninstall.destPath = fmt.Sprintf("%s/%s", ctrl.ctrlDir, procs.Uninstall.Name)

	ctrl.procs = *procs
	return nil
}

func (ctrl *Controller) copyScriptsToController() error {
	// Ensure SSH connection is established (no-op if already connected)
	if err := ctrl.ssh.Connect(); err != nil {
		return err
	}

	// Copy scripts to remote host
	for idx, script := range ctrl.procs.scriptNames {
		content := ctrl.procs.scriptContents[idx]
		reader := strings.NewReader(content)
		if err := ctrl.ssh.CopyTo(reader, ctrl.ctrlDir, script, "0775", int64(len(content))); err != nil {
			return err
		}
	}
	return nil
}

func (ctrl *Controller) CopyScript(srcDir, filename, destDir string) (err error) {
	// Read script from assets
	if srcDir != "" {
		srcDir = util.AddTrailingSlash(srcDir)
	}
	staticFile, err := util.GetStaticFile(srcDir + filename)
	if err != nil {
		return err
	}

	// Copy to /tmp for backwards compatibility
	reader := strings.NewReader(staticFile)
	if err := ctrl.ssh.CopyTo(reader, destDir, filename, "0775", int64(len(staticFile))); err != nil {
		return err
	}

	return nil
}

func (ctrl *Controller) Uninstall() (err error) {
	// Stop controller gracefully
	if err = ctrl.Stop(); err != nil {
		return
	}

	// Connect to server
	Verbose("Connecting to server")
	if err = ctrl.ssh.Connect(); err != nil {
		return
	}
	defer util.Log(ctrl.ssh.Disconnect)

	// Copy uninstallation scripts to remote host
	Verbose("Copying uninstall files to server")
	if _, err = ctrl.ssh.Run(fmt.Sprintf("sudo mkdir -p %s && sudo chmod -R 0777 %s/", ctrl.ctrlDir, ctrl.ctrlDir)); err != nil {
		return err
	}

	// Use custom scripts if available, otherwise use default embedded scripts
	if ctrl.customInstall || len(ctrl.procs.scriptNames) > 0 {
		// Copy uninstall script specifically
		uninstallIdx := -1
		for idx, scriptName := range ctrl.procs.scriptNames {
			if scriptName == pkg.controllerScriptUninstall {
				uninstallIdx = idx
				break
			}
		}
		if uninstallIdx >= 0 {
			content := ctrl.procs.scriptContents[uninstallIdx]
			reader := strings.NewReader(content)
			if err := ctrl.ssh.CopyTo(reader, ctrl.ctrlDir, pkg.controllerScriptUninstall, "0775", int64(len(content))); err != nil {
				return err
			}
		} else {
			// Fallback to default uninstall script
			assetPrefix := controllerAssetPrefixContainer
			if ctrl.Airgap {
				assetPrefix = controllerAssetPrefixAirgap
			}
			if err := ctrl.CopyScript(assetPrefix, pkg.controllerScriptUninstall, ctrl.ctrlDir); err != nil {
				return err
			}
		}
	} else {
		// Fallback to default method for backward compatibility
		assetPrefix := controllerAssetPrefixContainer
		if ctrl.Airgap {
			assetPrefix = controllerAssetPrefixAirgap
		}
		if err := ctrl.CopyScript(assetPrefix, pkg.controllerScriptUninstall, ctrl.ctrlDir); err != nil {
			return err
		}
	}

	// Use uninstall entrypoint if available
	uninstallCmd := ctrl.procs.Uninstall.getCommand()
	if uninstallCmd == "" {
		uninstallCmd = fmt.Sprintf("%s/%s", ctrl.ctrlDir, pkg.controllerScriptUninstall)
	}

	cmds := []command{
		{
			cmd: fmt.Sprintf("sudo %s", uninstallCmd),
			msg: "Uninstalling controller on host " + ctrl.Host,
		},
	}

	// Execute commands
	for _, cmd := range cmds {
		Verbose(cmd.msg)
		_, err = ctrl.ssh.Run(cmd.cmd)
		if err != nil {
			return
		}
	}
	return nil
}

func (ctrl *Controller) copyInstallScripts() error {
	Verbose("Copying install files to server")
	if _, err := ctrl.ssh.Run(fmt.Sprintf("sudo mkdir -p %s && sudo chmod -R 0777 %s/", ctrl.ctrlDir, ctrl.ctrlDir)); err != nil {
		return err
	}

	// Use custom scripts if available, otherwise use default embedded scripts
	if ctrl.customInstall || len(ctrl.procs.scriptNames) > 0 {
		return ctrl.copyScriptsToController()
	}

	// Fallback to default method for backward compatibility
	assetPrefix := controllerAssetPrefixContainer
	if ctrl.Airgap {
		assetPrefix = controllerAssetPrefixAirgap
	}
	scripts := []string{
		pkg.controllerScriptPrereq,
		pkg.controllerScriptInit,
		pkg.controllerScriptInstallContainerEngine,
		pkg.controllerScriptInstall,
		pkg.controllerScriptSetEnv,
	}
	for _, script := range scripts {
		if err := ctrl.CopyScript(assetPrefix, script, ctrl.ctrlDir); err != nil {
			return err
		}
	}
	return nil
}

func appendControllerBaseEnv(env []string, ctrl *Controller) []string {
	env = append(env, "CONTROL_PLANE=Remote")
	env = append(env, fmt.Sprintf("\"CONTROLLER_NAMESPACE=%s\"", ctrl.Namespace))
	if ctrl.Https != nil && ctrl.Https.Enabled != nil && *ctrl.Https.Enabled {
		env = append(env, fmt.Sprintf("\"SERVER_DEV_MODE=%s\"", "false"))
		env = append(env, fmt.Sprintf("\"SSL_BASE64_CERT=%s\"", ctrl.Https.TLSCert))
		env = append(env, fmt.Sprintf("\"SSL_BASE64_KEY=%s\"", ctrl.Https.TLSKey))
	}
	if ctrl.Https != nil && ctrl.Https.CACert != "" {
		env = append(env, fmt.Sprintf("\"SSL_BASE64_INTERMEDIATE_CERT=%s\"", ctrl.Https.CACert))
	}
	if ctrl.Host != "" {
		env = append(env, fmt.Sprintf(`"CONTROLLER_HOST=%s"`, ctrl.Host))
	}
	if ctrl.PidBaseDir != "" {
		env = append(env, fmt.Sprintf("\"PID_BASE=%s\"", ctrl.PidBaseDir))
	}
	if ctrl.EcnViewerPort != 0 {
		env = append(env, fmt.Sprintf("\"VIEWER_PORT=%d\"", ctrl.EcnViewerPort))
	}
	if ctrl.EcnViewerURL != "" {
		env = append(env, fmt.Sprintf("\"VIEWER_URL=%s\"", ctrl.EcnViewerURL))
	}
	if ctrl.SystemMicroservices.Router.AMD64 != "" {
		env = append(env, fmt.Sprintf("\"ROUTER_IMAGE_1=%s\"", ctrl.SystemMicroservices.Router.AMD64))
	}
	if ctrl.SystemMicroservices.Router.ARM64 != "" {
		env = append(env, fmt.Sprintf("\"ROUTER_IMAGE_2=%s\"", ctrl.SystemMicroservices.Router.ARM64))
	}
	if ctrl.SystemMicroservices.Router.RISCV64 != "" {
		env = append(env, fmt.Sprintf("\"ROUTER_IMAGE_3=%s\"", ctrl.SystemMicroservices.Router.RISCV64))
	}
	if ctrl.SystemMicroservices.Router.ARM != "" {
		env = append(env, fmt.Sprintf("\"ROUTER_IMAGE_4=%s\"", ctrl.SystemMicroservices.Router.ARM))
	}
	return env
}

func appendDBEnv(env []string, ctrl *Controller) []string {
	if ctrl.db.host != "" {
		env = append(env,
			fmt.Sprintf(`"DB_PROVIDER=%s"`, ctrl.db.provider),
			fmt.Sprintf(`"DB_HOST=%s"`, ctrl.db.host),
			fmt.Sprintf(`"DB_USERNAME=%s"`, ctrl.db.user),
			fmt.Sprintf(`"DB_PASSWORD=%s"`, ctrl.db.password),
			fmt.Sprintf(`"DB_PORT=%d"`, ctrl.db.port),
			fmt.Sprintf(`"DB_NAME=%s"`, ctrl.db.databaseName))
	}
	if ctrl.db.ssl != nil {
		env = append(env, fmt.Sprintf(`"DB_USE_SSL=%t"`, *ctrl.db.ssl))
	}
	if ctrl.db.ca != nil {
		env = append(env, fmt.Sprintf(`"DB_SSL_CA=%s"`, *ctrl.db.ca))
	}
	return env
}

func appendAuthEnv(env []string, ctrl *Controller) []string {
	if ctrl.auth.url != "" {
		env = append(env,
			fmt.Sprintf(`"KC_URL=%s"`, ctrl.auth.url),
			fmt.Sprintf(`"KC_REALM=%s"`, ctrl.auth.realm),
			fmt.Sprintf(`"KC_SSL_REQ=%s"`, ctrl.auth.ssl),
			fmt.Sprintf(`"KC_REALM_KEY=%s"`, ctrl.auth.realmKey),
			fmt.Sprintf(`"KC_CLIENT=%s"`, ctrl.auth.controllerClient),
			fmt.Sprintf(`"KC_CLIENT_SECRET=%s"`, ctrl.auth.controllerSecret),
			fmt.Sprintf(`"KC_VIEWER_CLIENT=%s"`, ctrl.auth.viewerClient))
	}
	return env
}

func appendNatsEnv(env []string, ctrl *Controller) []string {
	natsAMD64 := ctrl.SystemMicroservices.Nats.AMD64
	natsARM64 := ctrl.SystemMicroservices.Nats.ARM64
	natsRISCV64 := ctrl.SystemMicroservices.Nats.RISCV64
	natsARM := ctrl.SystemMicroservices.Nats.ARM
	if natsAMD64 == "" && natsARM64 != "" {
		natsAMD64 = natsARM64
	}
	if natsARM64 == "" && natsAMD64 != "" {
		natsARM64 = natsAMD64
	}
	if natsRISCV64 == "" && natsARM != "" {
		natsRISCV64 = natsARM
	}
	if natsARM == "" && natsRISCV64 != "" {
		natsARM = natsRISCV64
	}
	if natsAMD64 == "" && natsARM64 == "" && natsRISCV64 == "" {
		natsAMD64 = util.GetNatsImage()
	}
	if natsAMD64 != "" {
		env = append(env, fmt.Sprintf("\"NATS_IMAGE_1=%s\"", natsAMD64))
	}
	if natsARM64 != "" {
		env = append(env, fmt.Sprintf("\"NATS_IMAGE_2=%s\"", natsARM64))
	}
	if natsRISCV64 != "" {
		env = append(env, fmt.Sprintf("\"NATS_IMAGE_3=%s\"", natsRISCV64))
	}
	if natsARM != "" {
		env = append(env, fmt.Sprintf("\"NATS_IMAGE_4=%s\"", natsARM))
	}
	natsEnabled := true
	if ctrl.NatsEnabled != nil {
		natsEnabled = *ctrl.NatsEnabled
	}
	env = append(env, fmt.Sprintf("\"NATS_ENABLED=%t\"", natsEnabled))
	return env
}

func appendVaultEnv(env []string, ctrl *Controller) []string {
	if ctrl.Vault == nil {
		return env
	}
	if ctrl.Vault.Enabled != nil {
		env = append(env, fmt.Sprintf("\"VAULT_ENABLED=%t\"", *ctrl.Vault.Enabled))
	}
	if ctrl.Vault.Provider != "" {
		env = append(env, fmt.Sprintf("\"VAULT_PROVIDER=%s\"", ctrl.Vault.Provider))
	}
	if ctrl.Vault.BasePath != "" {
		env = append(env, fmt.Sprintf("\"VAULT_BASE_PATH=%s\"", ctrl.Vault.BasePath))
	}
	env = appendVaultHashicorpEnv(env, ctrl.Vault.Hashicorp)
	env = appendVaultAwsEnv(env, ctrl.Vault.Aws)
	env = appendVaultAzureEnv(env, ctrl.Vault.Azure)
	env = appendVaultGoogleEnv(env, ctrl.Vault.Google)
	return env
}

func appendVaultHashicorpEnv(env []string, h *VaultHashicorpConfig) []string {
	if h == nil {
		return env
	}
	if h.Address != "" {
		env = append(env, fmt.Sprintf("\"VAULT_HASHICORP_ADDRESS=%s\"", h.Address))
	}
	if h.Token != "" {
		env = append(env, fmt.Sprintf("\"VAULT_HASHICORP_TOKEN=%s\"", h.Token))
	}
	if h.Mount != "" {
		env = append(env, fmt.Sprintf("\"VAULT_HASHICORP_MOUNT=%s\"", h.Mount))
	}
	return env
}

func appendVaultAwsEnv(env []string, a *VaultAwsConfig) []string {
	if a == nil {
		return env
	}
	if a.Region != "" {
		env = append(env, fmt.Sprintf("\"VAULT_AWS_REGION=%s\"", a.Region))
	}
	if a.AccessKeyId != "" {
		env = append(env, fmt.Sprintf("\"VAULT_AWS_ACCESS_KEY_ID=%s\"", a.AccessKeyId))
	}
	if a.AccessKey != "" {
		env = append(env, fmt.Sprintf("\"VAULT_AWS_ACCESS_KEY=%s\"", a.AccessKey))
	}
	return env
}

func appendVaultAzureEnv(env []string, a *VaultAzureConfig) []string {
	if a == nil {
		return env
	}
	if a.URL != "" {
		env = append(env, fmt.Sprintf("\"VAULT_AZURE_URL=%s\"", a.URL))
	}
	if a.TenantId != "" {
		env = append(env, fmt.Sprintf("\"VAULT_AZURE_TENANT_ID=%s\"", a.TenantId))
	}
	if a.ClientId != "" {
		env = append(env, fmt.Sprintf("\"VAULT_AZURE_CLIENT_ID=%s\"", a.ClientId))
	}
	if a.ClientSecret != "" {
		env = append(env, fmt.Sprintf("\"VAULT_AZURE_CLIENT_SECRET=%s\"", a.ClientSecret))
	}
	return env
}

func appendVaultGoogleEnv(env []string, g *VaultGoogleConfig) []string {
	if g == nil {
		return env
	}
	if g.ProjectId != "" {
		env = append(env, fmt.Sprintf("\"VAULT_GOOGLE_PROJECT_ID=%s\"", g.ProjectId))
	}
	if g.Credentials != "" {
		env = append(env, fmt.Sprintf("\"VAULT_GOOGLE_CREDENTIALS=%s\"", g.Credentials))
	}
	return env
}

func appendLogLevelEnv(env []string, ctrl *Controller) []string {
	if ctrl.LogLevel != "" {
		env = append(env, fmt.Sprintf("\"LOG_LEVEL=%s\"", ctrl.LogLevel))
	}
	return env
}

func appendEventsEnv(env []string, ctrl *Controller) []string {
	if ctrl.events.auditEnabled == nil {
		return env
	}
	env = append(env, fmt.Sprintf("\"EVENT_AUDIT_ENABLED=%t\"", *ctrl.events.auditEnabled))
	if *ctrl.events.auditEnabled {
		if ctrl.events.retentionDays != 0 {
			env = append(env, fmt.Sprintf("\"EVENT_RETENTION_DAYS=%d\"", ctrl.events.retentionDays))
		}
		if ctrl.events.cleanupInterval != 0 {
			env = append(env, fmt.Sprintf("\"EVENT_CLEANUP_INTERVAL=%d\"", ctrl.events.cleanupInterval))
		}
	}
	if ctrl.events.captureIpAddress != nil {
		env = append(env, fmt.Sprintf("\"EVENT_CAPTURE_IP_ADDRESS=%t\"", *ctrl.events.captureIpAddress))
	}
	return env
}

func (ctrl *Controller) prepareEnvironmentVariables() string {
	env := []string{}
	env = appendControllerBaseEnv(env, ctrl)
	env = appendDBEnv(env, ctrl)
	env = appendAuthEnv(env, ctrl)
	env = appendNatsEnv(env, ctrl)
	env = appendVaultEnv(env, ctrl)
	env = appendLogLevelEnv(env, ctrl)
	env = appendEventsEnv(env, ctrl)
	return strings.Join(env, " ")
}

func (ctrl *Controller) prepareCommands(envString string) []command {
	// Define commands - use custom entrypoints if available
	checkPrereqsCmd := ctrl.procs.check.getCommand()
	if checkPrereqsCmd == "" {
		checkPrereqsCmd = fmt.Sprintf("%s/%s", ctrl.ctrlDir, pkg.controllerScriptPrereq)
	}
	depsCmd := ctrl.procs.Deps.getCommand()
	if depsCmd == "" {
		depsCmd = fmt.Sprintf("sudo %s/%s", ctrl.ctrlDir, pkg.controllerScriptInstallContainerEngine)
	} else {
		depsCmd = fmt.Sprintf("sudo %s", depsCmd)
	}
	setEnvCmd := ctrl.procs.SetEnv.getCommand()
	if setEnvCmd == "" {
		setEnvCmd = fmt.Sprintf("sudo %s/%s %s", ctrl.ctrlDir, pkg.controllerScriptSetEnv, envString)
	} else {
		setEnvCmd = fmt.Sprintf("sudo %s %s", setEnvCmd, envString)
	}
	installCmd := ctrl.procs.Install.getCommand()
	if installCmd == "" {
		installCmd = fmt.Sprintf("sudo %s/%s %s", ctrl.ctrlDir, pkg.controllerScriptInstall, ctrl.Image)
	} else {
		installCmd = fmt.Sprintf("sudo %s", installCmd)
	}

	return []command{
		{
			cmd: checkPrereqsCmd,
			msg: "Checking prerequisites on Controller " + ctrl.Host,
		},
		{
			cmd: depsCmd,
			msg: "Installing dependencies on Controller " + ctrl.Host,
		},
		{
			cmd: setEnvCmd,
			msg: "Setting up environment variables for Controller " + ctrl.Host,
		},
		{
			cmd: installCmd,
			msg: "Installing ioFog on Controller " + ctrl.Host,
		},
	}
}

func (ctrl *Controller) executeCommands(cmds []command) error {
	for _, cmd := range cmds {
		Verbose(cmd.msg)
		_, err := ctrl.ssh.Run(cmd.cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ctrl *Controller) waitForControllerToStart() (string, error) {
	// Specify errors to ignore while waiting
	ignoredErrors := []string{
		"Process exited with status 7", // curl: (7) Failed to connect to localhost port 8080: Connection refused
	}

	// Add a small delay before checking
	time.Sleep(5 * time.Second)

	Verbose("Waiting for Controller " + ctrl.Host)
	// Increase timeout or retry attempts
	maxRetries := 15
	var protocol string
	if ctrl.Https != nil && ctrl.Https.Enabled != nil && *ctrl.Https.Enabled {
		protocol = "https"
	} else {
		protocol = "http"
	}
	for i := 0; i < maxRetries; i++ {
		Verbose(fmt.Sprintf("Try %d of %d", i+1, maxRetries))
		err := ctrl.ssh.RunUntil(
			regexp.MustCompile("\"status\":\"online\""),
			fmt.Sprintf("curl --request GET --url %s://localhost:%s/api/v3/status", protocol, iofog.ControllerPortString),
			ignoredErrors,
		)
		if err == nil {
			return protocol, nil
		}
		time.Sleep(2 * time.Second * time.Duration(i+1)) // Exponential backoff
	}
	return "", fmt.Errorf("controller failed to start after %d retries", maxRetries)
}

func (ctrl *Controller) Install() (err error) {
	// Connect to server
	Verbose("Connecting to server")
	if err = ctrl.ssh.Connect(); err != nil {
		return
	}
	defer util.Log(ctrl.ssh.Disconnect)

	// Copy installation scripts to remote host
	if err = ctrl.copyInstallScripts(); err != nil {
		return err
	}

	// Prepare and build environment variables
	envString := ctrl.prepareEnvironmentVariables()

	// Prepare commands
	cmds := ctrl.prepareCommands(envString)

	// Execute commands
	if err = ctrl.executeCommands(cmds); err != nil {
		return err
	}

	// Wait for controller to start
	protocol, err := ctrl.waitForControllerToStart()
	if err != nil {
		return err
	}

	// Wait for API
	endpoint := fmt.Sprintf("%s://%s:%s", protocol, ctrl.Host, iofog.ControllerPortString)
	if err = WaitForControllerAPI(endpoint); err != nil {
		return
	}

	return nil
}

func (ctrl *Controller) Stop() (err error) {
	// Connect to server
	if err = ctrl.ssh.Connect(); err != nil {
		return
	}
	defer util.Log(ctrl.ssh.Disconnect)

	// TODO: Clear the database
	// Define commands
	cmds := []string{
		"sudo service iofog-controller stop",
	}

	// Execute commands
	for _, cmd := range cmds {
		_, err = ctrl.ssh.Run(cmd)
		if err != nil {
			return
		}
	}

	return
}

func WaitForControllerAPI(endpoint string) (err error) {
	baseURL, err := util.GetBaseURL(endpoint)
	if err != nil {
		return err
	}
	ctrlClient := client.New(client.Options{BaseURL: baseURL})

	seconds := 0
	for seconds < 60 {
		// Try to create the user, return if success
		if _, err = ctrlClient.GetStatus(); err == nil {
			return
		}
		// Connection failed, wait and retry
		time.Sleep(time.Millisecond * 1000)
		seconds++
	}

	// Return last error
	return
}

func DeployRouterSecrets(endpoint, secretName string, TLSCert, TLSKey string) (err error) {
	baseURL, err := util.GetBaseURL(endpoint)
	if err != nil {
		return err
	}
	ctrlClient := client.New(client.Options{BaseURL: baseURL})

	request := client.SecretCreateRequest{
		Name: secretName,
		Type: "tls",
		Data: map[string]string{
			"ca.crt":  TLSCert,
			"tls.crt": TLSCert,
			"tls.key": TLSKey,
		},
	}

	if err = ctrlClient.CreateSecret(&request); err != nil {
		return err
	}
	return nil
}

func ImportRouterCertificate(endpoint string, secretName string) (err error) {
	baseURL, err := util.GetBaseURL(endpoint)
	if err != nil {
		return err
	}
	ctrlClient := client.New(client.Options{BaseURL: baseURL})

	// Create CA certificate
	request := client.CACreateRequest{
		Name:       secretName,
		Type:       "direct",
		SecretName: secretName,
	}

	if err = ctrlClient.CreateCA(&request); err != nil {
		return err
	}

	return nil
}

// GlobalCertificates holds optional router/NATS CA blocks deployed once via Controller API.
type GlobalCertificates struct {
	RouterSiteCA  *SiteCertificate
	RouterLocalCA *SiteCertificate
	NatsSiteCA    *SiteCertificate
	NatsLocalCA   *SiteCertificate
}

func (g GlobalCertificates) empty() bool {
	return g.RouterSiteCA == nil && g.RouterLocalCA == nil && g.NatsSiteCA == nil && g.NatsLocalCA == nil
}

// DeployGlobalCertificates uploads router and NATS site/local CA secrets once (authenticated client).
func DeployGlobalCertificates(clt *client.Client, certs GlobalCertificates) error {
	if clt == nil || certs.empty() {
		return nil
	}
	pairs := []struct {
		name string
		cert *SiteCertificate
	}{
		{"router-site-ca", certs.RouterSiteCA},
		{"default-router-local-ca", certs.RouterLocalCA},
		{"nats-site-ca", certs.NatsSiteCA},
		{"default-nats-local-ca", certs.NatsLocalCA},
	}
	for _, pair := range pairs {
		if err := deployGlobalCertificate(clt, pair.name, pair.cert); err != nil {
			return err
		}
	}
	return nil
}

type globalCertDeployer interface {
	GetSecret(name string) (*client.SecretInfo, error)
	GetCA(name string) (*client.CAInfo, error)
	CreateSecret(request *client.SecretCreateRequest) error
	CreateCA(request *client.CACreateRequest) error
}

func isControllerNotFound(err error) bool {
	var notFound *client.NotFoundError
	return errors.As(err, &notFound)
}

func isControllerConflict(err error) bool {
	var conflict *client.ConflictError
	return errors.As(err, &conflict)
}

func controllerSecretExists(clt globalCertDeployer, name string) (bool, error) {
	_, err := clt.GetSecret(name)
	if err == nil {
		return true, nil
	}
	if isControllerNotFound(err) {
		return false, nil
	}
	return false, err
}

func controllerCAExists(clt globalCertDeployer, name string) (bool, error) {
	_, err := clt.GetCA(name)
	if err == nil {
		return true, nil
	}
	if isControllerNotFound(err) {
		return false, nil
	}
	return false, err
}

func deployGlobalCertificate(clt globalCertDeployer, secretName string, cert *SiteCertificate) error {
	if cert == nil {
		return nil
	}

	secretExists, err := controllerSecretExists(clt, secretName)
	if err != nil {
		return err
	}
	caExists, err := controllerCAExists(clt, secretName)
	if err != nil {
		return err
	}
	if secretExists && caExists {
		return nil
	}

	if !secretExists {
		secretRequest := client.SecretCreateRequest{
			Name: secretName,
			Type: "tls",
			Data: map[string]string{
				"ca.crt":  cert.TLSCert,
				"tls.crt": cert.TLSCert,
				"tls.key": cert.TLSKey,
			},
		}
		if err := clt.CreateSecret(&secretRequest); err != nil && !isControllerConflict(err) {
			return err
		}
	}

	if !caExists {
		caRequest := client.CACreateRequest{
			Name:       secretName,
			Type:       "direct",
			SecretName: secretName,
		}
		if err := clt.CreateCA(&caRequest); err != nil && !isControllerConflict(err) {
			return err
		}
	}
	return nil
}
