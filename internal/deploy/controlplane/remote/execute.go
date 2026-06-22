package deployremotecontrolplane

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	deployagent "github.com/eclipse-iofog/iofogctl/internal/deploy/agent"
	deployagentconfig "github.com/eclipse-iofog/iofogctl/internal/deploy/agentconfig"
	deployairgap "github.com/eclipse-iofog/iofogctl/internal/deploy/airgap"
	deployremotecontroller "github.com/eclipse-iofog/iofogctl/internal/deploy/controller/remote"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	iutil "github.com/eclipse-iofog/iofogctl/internal/util"

	// clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"

	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

const (
	deploymentTypeContainer = "container"
	deploymentTypeNative    = "native"

	defaultNatsServerPort    = 4222
	defaultNatsClusterPort   = 6222
	defaultNatsLeafPort      = 7422
	defaultNatsMqttPort      = 8883
	defaultNatsHttpPort      = 8222
	defaultJsStorageSize     = "10G"
	defaultJsMemoryStoreSize = "1G"
)

// applySystemAgentNatsDefaults sets NATS config defaults for system agents when not provided (natsMode=server, ports, JsStorageSize, JsMemoryStoreSize).
func applySystemAgentNatsDefaults(cfg *rsc.AgentConfiguration) {
	if cfg.NatsMode == nil {
		cfg.NatsMode = iutil.MakeStrPtr(iofog.NatsModeServer)
	} else {
		// Force server mode for system agents (like router interior)
		cfg.NatsMode = iutil.MakeStrPtr(iofog.NatsModeServer)
	}
	if cfg.NatsServerPort == nil {
		cfg.NatsServerPort = iutil.MakeIntPtr(defaultNatsServerPort)
	}
	if cfg.NatsClusterPort == nil {
		cfg.NatsClusterPort = iutil.MakeIntPtr(defaultNatsClusterPort)
	}
	if cfg.NatsLeafPort == nil {
		cfg.NatsLeafPort = iutil.MakeIntPtr(defaultNatsLeafPort)
	}
	if cfg.NatsMqttPort == nil {
		cfg.NatsMqttPort = iutil.MakeIntPtr(defaultNatsMqttPort)
	}
	if cfg.NatsHTTPPort == nil {
		cfg.NatsHTTPPort = iutil.MakeIntPtr(defaultNatsHttpPort)
	}
	if cfg.JsStorageSize == nil {
		cfg.JsStorageSize = iutil.MakeStrPtr(defaultJsStorageSize)
	}
	if cfg.JsMemoryStoreSize == nil {
		cfg.JsMemoryStoreSize = iutil.MakeStrPtr(defaultJsMemoryStoreSize)
	}
}

type Options struct {
	Namespace string
	Yaml      []byte
	Name      string
}

type remoteControlPlaneExecutor struct {
	ctrlClient          *client.Client
	controllerExecutors []execute.Executor
	controlPlane        rsc.ControlPlane
	ns                  *rsc.Namespace
	name                string
}

func deploySystemAgent(namespace string, ctrl *rsc.RemoteController, systemAgentConfig *rsc.SystemAgentConfig) (err error) {
	// Deploy system agent to host internal router
	install.Verbose("Deploying system agent for controller " + ctrl.Name)
	// If DeploymentType is nil, default to "container"
	var deploymentType string
	if systemAgentConfig != nil && systemAgentConfig.AgentConfiguration != nil && systemAgentConfig.AgentConfiguration.DeploymentType != nil {
		// Use DeploymentType from provided configuration
		deploymentType = *systemAgentConfig.AgentConfiguration.DeploymentType
	} else if systemAgentConfig != nil && systemAgentConfig.Package.Container.Image != "" {
		// If container image is specified, use container
		deploymentType = deploymentTypeContainer
	} else {
		// Default to container if DeploymentType is nil
		deploymentType = deploymentTypeContainer
	}

	// Get agent configuration - use provided config or defaults
	var deployAgentConfig rsc.AgentConfiguration
	if systemAgentConfig != nil && systemAgentConfig.AgentConfiguration != nil {
		// Use provided configuration
		deployAgentConfig = *systemAgentConfig.AgentConfiguration
		// Ensure host is set
		if deployAgentConfig.Host == nil {
			deployAgentConfig.Host = &ctrl.Host
		}
		// Ensure IsSystem is always true for system agents
		deployAgentConfig.IsSystem = iutil.MakeBoolPtr(true)
		// Ensure DeploymentType is set (default to container if nil)
		if deployAgentConfig.DeploymentType == nil {
			deployAgentConfig.DeploymentType = iutil.MakeStrPtr(deploymentType)
		}
	} else {
		// Use defaults with configurable ports (router mode always interior)
		RouterConfig := client.RouterConfig{
			RouterMode:      iutil.MakeStrPtr(iofog.RouterModeInterior),
			MessagingPort:   iutil.MakeIntPtr(5671),
			EdgeRouterPort:  iutil.MakeIntPtr(45671),
			InterRouterPort: iutil.MakeIntPtr(55671),
		}

		upstreamRouters := []string{}
		upstreamNatsServers := []string{}

		deployAgentConfig = rsc.AgentConfiguration{
			Name: ctrl.Name,
			Arch: iutil.MakeStrPtr("auto"),
			AgentConfiguration: client.AgentConfiguration{
				IsSystem:            iutil.MakeBoolPtr(true),
				DeploymentType:      iutil.MakeStrPtr(deploymentType),
				Host:                &ctrl.Host,
				RouterConfig:        RouterConfig,
				UpstreamRouters:     &upstreamRouters,
				UpstreamNatsServers: &upstreamNatsServers,
			},
		}
	}

	// Ensure router mode is always "interior" for system agents
	if deployAgentConfig.RouterMode == nil {
		interior := iofog.RouterModeInterior
		deployAgentConfig.RouterMode = &interior
	} else if *deployAgentConfig.RouterMode != iofog.RouterModeInterior {
		// Force to interior mode
		interior := iofog.RouterModeInterior
		deployAgentConfig.RouterMode = &interior
	}

	if deployAgentConfig.EdgeRouterPort == nil {
		edgeRouterPort := 45671
		deployAgentConfig.EdgeRouterPort = &edgeRouterPort
	}
	if deployAgentConfig.InterRouterPort == nil {
		interRouterPort := 55671
		deployAgentConfig.InterRouterPort = &interRouterPort
	}

	if deployAgentConfig.MessagingPort == nil {
		messagingPort := 5671
		deployAgentConfig.MessagingPort = &messagingPort
	}

	// System agents run NATS in server mode (like router interior). Apply default natsConfig when not provided.
	applySystemAgentNatsDefaults(&deployAgentConfig)

	// Ensure name is set
	if deployAgentConfig.Name == "" {
		deployAgentConfig.Name = ctrl.Name
	}

	agent := rsc.RemoteAgent{
		Name:   ctrl.Name,
		Host:   ctrl.Host,
		SSH:    ctrl.SSH,
		Config: &deployAgentConfig,
	}
	// Set Package and Scripts if systemAgentConfig is provided
	if systemAgentConfig != nil {
		agent.Package = systemAgentConfig.Package
		agent.Scripts = systemAgentConfig.Scripts // Support custom scripts
	}

	// Get Agentconfig executor
	deployAgentConfigExecutor := deployagentconfig.NewRemoteExecutor(ctrl.Name, &deployAgentConfig, namespace, nil)
	// If there already is a system fog, ignore error
	if err := deployAgentConfigExecutor.Execute(); err != nil {
		return err
	}
	agent.UUID = deployAgentConfigExecutor.GetAgentUUID()
	agentDeployExecutor, err := deployagent.NewRemoteExecutor(namespace, &agent, true) // isSystem = true
	if err != nil {
		return err
	}
	return agentDeployExecutor.Execute()
}

func deployNextSystemAgent(namespace string, ctrl *rsc.RemoteController, systemAgentConfig *rsc.SystemAgentConfig) (err error) {
	// Deploy system agent to host internal router
	install.Verbose("Deploying next-system agent for controller " + ctrl.Name)
	// If DeploymentType is nil, default to "container"
	var deploymentType string
	if systemAgentConfig != nil && systemAgentConfig.AgentConfiguration != nil && systemAgentConfig.AgentConfiguration.DeploymentType != nil {
		// Use DeploymentType from provided configuration
		deploymentType = *systemAgentConfig.AgentConfiguration.DeploymentType
	} else if systemAgentConfig != nil && systemAgentConfig.Package.Container.Image != "" {
		// If container image is specified, use container
		deploymentType = deploymentTypeContainer
	} else {
		// Default to container if DeploymentType is nil
		deploymentType = deploymentTypeContainer
	}

	// Get agent configuration - use provided config or defaults
	var deployAgentConfig rsc.AgentConfiguration
	if systemAgentConfig != nil && systemAgentConfig.AgentConfiguration != nil {
		// Use provided configuration
		deployAgentConfig = *systemAgentConfig.AgentConfiguration
		// Ensure host is set
		if deployAgentConfig.Host == nil {
			deployAgentConfig.Host = &ctrl.Host
		}
		// Ensure IsSystem is always true for system agents
		deployAgentConfig.IsSystem = iutil.MakeBoolPtr(true)
		// Ensure DeploymentType is set (default to container if nil)
		if deployAgentConfig.DeploymentType == nil {
			deployAgentConfig.DeploymentType = iutil.MakeStrPtr(deploymentType)
		}
		// Override upstream routers for non-first controllers
		if deployAgentConfig.UpstreamRouters == nil {
			upstreamRouters := []string{"default-router"}
			deployAgentConfig.UpstreamRouters = &upstreamRouters
		} else {
			// Add default-router if not already present
			hasDefaultRouter := false
			for _, router := range *deployAgentConfig.UpstreamRouters {
				if router == "default-router" {
					hasDefaultRouter = true
					break
				}
			}
			if !hasDefaultRouter {
				*deployAgentConfig.UpstreamRouters = append(*deployAgentConfig.UpstreamRouters, "default-router")
			}
		}
		// Override upstream nats server for non-first controllers
		if deployAgentConfig.UpstreamNatsServers == nil {
			upstreamNatsServers := []string{"default-nats-hub"}
			deployAgentConfig.UpstreamNatsServers = &upstreamNatsServers
		} else {
			// Add default-nats-hub if not already present
			hasDefaultNatsHub := false
			for _, natsServer := range *deployAgentConfig.UpstreamNatsServers {
				if natsServer == "default-nats-hub" {
					hasDefaultNatsHub = true
					break
				}
			}
			if !hasDefaultNatsHub {
				*deployAgentConfig.UpstreamNatsServers = append(*deployAgentConfig.UpstreamNatsServers, "default-nats-hub")
			}
		}
	} else {
		// Use defaults with configurable ports (router mode always interior)
		RouterConfig := client.RouterConfig{
			RouterMode:      iutil.MakeStrPtr(iofog.RouterModeInterior),
			MessagingPort:   iutil.MakeIntPtr(5671),
			EdgeRouterPort:  iutil.MakeIntPtr(45671),
			InterRouterPort: iutil.MakeIntPtr(55671),
		}

		upstreamRouters := []string{"default-router"}

		deployAgentConfig = rsc.AgentConfiguration{
			Name: ctrl.Name,
			Arch: iutil.MakeStrPtr("auto"),
			AgentConfiguration: client.AgentConfiguration{
				IsSystem:        iutil.MakeBoolPtr(true),
				DeploymentType:  iutil.MakeStrPtr(deploymentType),
				Host:            &ctrl.Host,
				RouterConfig:    RouterConfig,
				UpstreamRouters: &upstreamRouters,
			},
		}
	}

	// Ensure router mode is always "interior" for system agents
	if deployAgentConfig.RouterMode == nil {
		interior := iofog.RouterModeInterior
		deployAgentConfig.RouterMode = &interior
	} else if *deployAgentConfig.RouterMode != iofog.RouterModeInterior {
		// Force to interior mode
		interior := iofog.RouterModeInterior
		deployAgentConfig.RouterMode = &interior
	}
	if deployAgentConfig.EdgeRouterPort == nil {
		edgeRouterPort := 45671
		deployAgentConfig.EdgeRouterPort = &edgeRouterPort
	}
	if deployAgentConfig.InterRouterPort == nil {
		interRouterPort := 55671
		deployAgentConfig.InterRouterPort = &interRouterPort
	}

	if deployAgentConfig.MessagingPort == nil {
		messagingPort := 5671
		deployAgentConfig.MessagingPort = &messagingPort
	}

	// System agents run NATS in server mode (like router interior). Apply default natsConfig when not provided.
	applySystemAgentNatsDefaults(&deployAgentConfig)

	// Ensure name is set
	if deployAgentConfig.Name == "" {
		deployAgentConfig.Name = ctrl.Name
	}

	agent := rsc.RemoteAgent{
		Name:   ctrl.Name,
		Host:   ctrl.Host,
		SSH:    ctrl.SSH,
		Config: &deployAgentConfig,
	}
	// Set Package and Scripts if systemAgentConfig is provided
	if systemAgentConfig != nil {
		agent.Package = systemAgentConfig.Package
		agent.Scripts = systemAgentConfig.Scripts // Support custom scripts
	}
	// Set airgap flag from control plane (get it from namespace)
	ns, err := config.GetNamespace(namespace)
	if err == nil {
		if cp, err := ns.GetControlPlane(); err == nil {
			if remoteCP, ok := cp.(*rsc.RemoteControlPlane); ok {
				agent.Airgap = remoteCP.Airgap
			}
		}
	}

	// Get Agentconfig executor
	deployAgentConfigExecutor := deployagentconfig.NewRemoteExecutor(ctrl.Name, &deployAgentConfig, namespace, nil)
	// If there already is a system fog, ignore error
	if err := deployAgentConfigExecutor.Execute(); err != nil {
		return err
	}
	agent.UUID = deployAgentConfigExecutor.GetAgentUUID()
	agentDeployExecutor, err := deployagent.NewRemoteExecutor(namespace, &agent, true) // isSystem = true
	if err != nil {
		return err
	}
	return agentDeployExecutor.Execute()
}

// prepareViewerURL prepares the viewer URL from controller configuration or endpoint
func prepareViewerURL(endpoint string) (string, error) {

	// Otherwise, construct from endpoint using logic similar to view.go
	URL, err := url.Parse(endpoint)
	if err != nil || URL.Host == "" {
		URL, err = url.Parse("//" + endpoint)
		if err != nil {
			return "", fmt.Errorf("failed to parse endpoint: %w", err)
		}
	}

	if URL.Scheme == "" {
		URL.Scheme = "http"
	}

	host := ""
	if strings.Contains(URL.Host, ":") {
		host, _, err = net.SplitHostPort(URL.Host)
		if err != nil {
			return "", fmt.Errorf("failed to split host and port: %w", err)
		}
	} else {
		host = URL.Host
	}

	// Add port for localhost
	if util.IsLocalHost(host) {
		host = net.JoinHostPort(host, iofog.ControllerHostECNViewerPortString)
	}

	URL.Host = host
	return URL.String(), nil
}

// updateViewerClientRootURL is retired in v3.8 (Keycloak viewer client).
func updateViewerClientRootURL(_ *rsc.RemoteControlPlane, _ string) error {
	return nil
}

func tagControllerImage(ctrl *rsc.RemoteController, image string) (err error) {

	if image == "" {
		image = util.GetControllerImage()
	}

	// Connect
	ssh, err := util.NewSecureShellClient(ctrl.SSH.User, ctrl.Host, ctrl.SSH.KeyFile)
	if err != nil {
		return err
	}
	if err := ssh.Connect(); err != nil {
		return err
	}

	defer util.Log(ssh.Disconnect)

	cmds := []string{
		fmt.Sprintf(`echo "IOFOG_CONTROLLER_IMAGE=%s" | sudo tee -a "/etc/iofog/agent/iofog-agent.env" > /dev/null`, image),
		"sudo service iofog-agent restart",
	}

	// Execute commands
	for _, cmd := range cmds {
		_, err = ssh.Run(cmd)
		if err != nil {
			return
		}
	}

	return
}

func (exe remoteControlPlaneExecutor) postDeploy() (err error) {
	controllers := exe.controlPlane.GetControllers()
	remoteControlPlane, ok := exe.controlPlane.(*rsc.RemoteControlPlane)
	if !ok {
		return util.NewInternalError("Could not convert ControlPlane to Remote ControlPlane")
	}

	// Check if airgap is enabled for system agents
	if remoteControlPlane.Airgap {
		// Transfer images for system agents before deployment
		if err := exe.transferSystemAgentImages(); err != nil {
			return fmt.Errorf("failed to transfer airgap images for system agents: %w", err)
		}
	}

	// Deploy agents for each controller
	for idx, baseController := range controllers {
		controller, ok := baseController.(*rsc.RemoteController)
		if !ok {
			return util.NewInternalError("Could not convert Controller to Remote Controller")
		}

		// // System agent config is required per controller
		// if controller.SystemAgent == nil {
		// 	return fmt.Errorf("controller '%s' must have a systemAgent configuration", controller.Name)
		// }

		// First controller gets system agent(with default-router), others get next-system agents(with interior mode)
		if idx == 0 {
			if err := deploySystemAgent(exe.ns.Name, controller, controller.SystemAgent); err != nil {
				return fmt.Errorf("failed to deploy system agent for first controller: %w", err)
			}
		} else {
			if err := deployNextSystemAgent(exe.ns.Name, controller, controller.SystemAgent); err != nil {
				return fmt.Errorf("failed to deploy next-system agent for controller %d: %w", idx, err)
			}
		}
		var image string
		// Check if controller has custom install script args (highest priority)
		if controller.Scripts != nil && controller.Scripts.Install.Args != nil && len(controller.Scripts.Install.Args) > 0 {
			image = controller.Scripts.Install.Args[0]
		} else if remoteControlPlane.Package.Container.Image != "" {
			// Use image from control plane package
			image = remoteControlPlane.Package.Container.Image
		} else {
			// Default to standard controller image
			image = util.GetControllerImage()
		}
		// Tag controller image for all controllers
		if err := tagControllerImage(controller, image); err != nil {
			return fmt.Errorf("failed to tag controller image for controller %d: %w", idx, err)
		}
	}
	return nil
}

func (exe remoteControlPlaneExecutor) Execute() (err error) {
	util.SpinStart(fmt.Sprintf("Deploying controlplane %s", exe.GetName()))

	// Check if airgap is enabled
	remoteControlPlane, ok := exe.controlPlane.(*rsc.RemoteControlPlane)
	if ok && remoteControlPlane.Airgap {
		if err := deployairgap.ValidateControlPlaneAirgapRequirements(remoteControlPlane); err != nil {
			return err
		}
		// Transfer only controller image; router and debugger are transferred in system agent phase
		if err := exe.transferControllerImages(); err != nil {
			return fmt.Errorf("failed to transfer airgap images for controllers: %w", err)
		}
	}

	if err := runExecutors(exe.controllerExecutors); err != nil {
		return err
	}

	// Make sure Controller API is ready
	endpoint, err := exe.controlPlane.GetEndpoint()
	if err != nil {
		return
	}
	if err := install.WaitForControllerAPI(endpoint); err != nil {
		return err
	}
	// // Create new user
	// baseURL, err := util.GetBaseURL(endpoint)
	// if err != nil {
	// 	return err
	// }
	// exe.ctrlClient = client.New(client.Options{BaseURL: baseURL})
	// user := client.User(exe.controlPlane.GetUser())
	// user.Password = exe.controlPlane.GetUser().GetRawPassword()
	// if err = exe.ctrlClient.CreateUser(user); err != nil {
	// 	// If not error about account existing, fail
	// 	if !strings.Contains(err.Error(), "already an account associated") {
	// 		return err
	// 	}
	// 	// Try to log in
	// 	if err := exe.ctrlClient.Login(client.LoginRequest{
	// 		Email:    user.Email,
	// 		Password: user.Password,
	// 	}); err != nil {
	// 		return err
	// 	}
	// }
	// Update config
	exe.ns.SetControlPlane(exe.controlPlane)
	if err := config.Flush(); err != nil {
		return err
	}

	// Update viewer client root URL if auth is configured
	if ok {
		if err := updateViewerClientRootURL(remoteControlPlane, endpoint); err != nil {
			// Log error but don't fail deployment
			util.PrintInfo(fmt.Sprintf("Warning: Failed to update viewer client root URL: %v\n", err))
		}
	}

	// Post deploy steps
	return exe.postDeploy()
}

func (exe remoteControlPlaneExecutor) GetName() string {
	return exe.name
}

func newControlPlaneExecutor(executors []execute.Executor, namespace *rsc.Namespace, name string, controlPlane rsc.ControlPlane) execute.Executor {
	return remoteControlPlaneExecutor{
		controllerExecutors: executors,
		ns:                  namespace,
		controlPlane:        controlPlane,
		name:                name,
	}
}

// Validates database configuration for multi-controller setup
func validateMultiControllerDatabase(controlPlane *rsc.RemoteControlPlane) error {
	if len(controlPlane.Controllers) > 1 {
		db := controlPlane.Database
		if db.Provider == "" || db.Host == "" || db.DatabaseName == "" ||
			db.Password == "" || db.Port == 0 || db.User == "" {
			return util.NewInputError("When deploying multiple controllers, you must specify an external database configuration with all required fields (host, user, password, provider, databaseName, port)")
		}
	}
	return nil
}

// Validates HTTPS configuration for a single controller
func validateControllerHTTPS(controller *rsc.RemoteController) error {
	if controller.Https != nil && controller.Https.Enabled != nil && *controller.Https.Enabled {
		// HTTPS is enabled, validate required fields
		if controller.Https.TLSCert == "" || controller.Https.TLSKey == "" {
			return util.NewInputError("When HTTPS is enabled, you must provide TLS certificate and key")
		}
	}
	return nil
}

// Validates CA configuration for a controller
func validateControllerRouterCA(controller *rsc.RemoteController) error {
	if controller.SiteCA != nil {
		if controller.SiteCA.TLSCert == "" || controller.SiteCA.TLSKey == "" {
			return util.NewInputError("When SiteCA is configured, you must provide both TLS certificate and key")
		}
	}
	if controller.LocalCA != nil {
		if controller.LocalCA.TLSCert == "" || controller.LocalCA.TLSKey == "" {
			return util.NewInputError("When LocalCA is configured, you must provide both TLS certificate and key")
		}
	}
	return nil
}

// Validates HTTPS configuration across all controllers
func validateMultiControllerHTTPS(controlPlane *rsc.RemoteControlPlane) error {
	controllers := controlPlane.Controllers
	if len(controllers) <= 1 {
		return nil
	}

	// Check first controller's HTTPS config
	firstController := controllers[0]
	if firstController.Https != nil && firstController.Https.Enabled != nil && *firstController.Https.Enabled {
		// First controller has HTTPS enabled, validate all controllers
		for idx, controller := range controllers {
			if err := validateControllerHTTPS(&controller); err != nil {
				return fmt.Errorf("controller %d (%s): %w", idx, controller.Name, err)
			}
		}
	}
	return nil
}

// Validates CA configuration across all controllers
func validateMultiControllerRouterCA(controlPlane *rsc.RemoteControlPlane) error {
	controllers := controlPlane.Controllers
	if len(controllers) <= 1 {
		return nil
	}

	// Only first controller should have CA configuration
	firstController := controllers[0]
	if firstController.SiteCA != nil || firstController.LocalCA != nil {
		// Validate first controller's CA config
		if err := validateControllerRouterCA(&firstController); err != nil {
			return fmt.Errorf("first controller (%s): %w", firstController.Name, err)
		}

		// Check that other controllers don't have CA config
		for idx, controller := range controllers[1:] {
			if controller.SiteCA != nil || controller.LocalCA != nil {
				return fmt.Errorf("controller %d (%s): CA configuration should only be specified for the first controller", idx+1, controller.Name)
			}
		}
	}
	return nil
}

// // Validates that each controller has a systemAgent configuration
// func validateControllerSystemAgent(controlPlane *rsc.RemoteControlPlane) error {
// 	controllers := controlPlane.Controllers
// 	if len(controllers) == 0 {
// 		return util.NewInputError("Remote Control Plane must have at least one controller")
// 	}

// 	for idx, controller := range controllers {
// 		if controller.SystemAgent == nil {
// 			return fmt.Errorf("controller %d (%s): systemAgent configuration is required", idx, controller.Name)
// 		}
// 		// Validate systemAgent package is provided
// 		if controller.SystemAgent.Package.Container.Image == "" && controller.SystemAgent.Package.Version == "" && controller.SystemAgent.Scripts.Install.Args == nil {
// 			return fmt.Errorf("controller %d (%s): systemAgent must have either package.container.image or package.version or scripts.install.args specified", idx, controller.Name)
// 		}
// 	}
// 	return nil
// }

// Main validation function that orchestrates all validations
func validateMultiControllerConfig(controlPlane *rsc.RemoteControlPlane) error {
	// // Validate systemAgent configuration
	// if err := validateControllerSystemAgent(controlPlane); err != nil {
	// 	return err
	// }

	// Validate database configuration
	if err := validateMultiControllerDatabase(controlPlane); err != nil {
		return err
	}

	// Validate HTTPS configuration
	if err := validateMultiControllerHTTPS(controlPlane); err != nil {
		return err
	}

	// Validate CA configuration
	if err := validateMultiControllerRouterCA(controlPlane); err != nil {
		return err
	}

	// Validate Vault when set (provider and required provider fields)
	if err := validateRemoteVault(controlPlane); err != nil {
		return err
	}

	return nil
}

func validateRemoteVault(controlPlane *rsc.RemoteControlPlane) error {
	if controlPlane.Vault == nil {
		return nil
	}
	v := controlPlane.Vault
	if v.Provider == "" {
		return nil
	}
	switch v.Provider {
	case "hashicorp", "openbao", "vault":
		if v.Hashicorp == nil || (v.Hashicorp.Address == "" && v.Hashicorp.Token == "") {
			return util.NewInputError("Vault provider " + v.Provider + " requires hashicorp block with address and token")
		}
	case "aws", "aws-secrets-manager":
		if v.Aws == nil {
			return util.NewInputError("Vault provider " + v.Provider + " requires aws block")
		}
	case "azure", "azure-key-vault":
		if v.Azure == nil {
			return util.NewInputError("Vault provider " + v.Provider + " requires azure block")
		}
	case "google", "google-secret-manager":
		if v.Google == nil {
			return util.NewInputError("Vault provider " + v.Provider + " requires google block")
		}
	}
	return nil
}

// transferControllerImages transfers only the controller image for airgap deployment.
// Router and debugger are transferred in the system agent phase (transferSystemAgentImages).
func (exe remoteControlPlaneExecutor) transferControllerImages() error {
	remoteControlPlane, ok := exe.controlPlane.(*rsc.RemoteControlPlane)
	if !ok {
		return util.NewInternalError("Could not convert ControlPlane to Remote ControlPlane")
	}

	isInitial, err := deployairgap.IsInitialDeployment(exe.ns.Name)
	if err != nil {
		return fmt.Errorf("failed to determine deployment type: %w", err)
	}

	images, err := deployairgap.CollectControllerImages(exe.ns.Name, remoteControlPlane, isInitial)
	if err != nil {
		return fmt.Errorf("failed to collect controller images: %w", err)
	}

	// Transfer controller and NATS images (remote Controller runs/starts NATS).
	// Use platform and container engine from system agent config (validated when airgap is enabled).
	imageList := []string{images.Controller}
	if images.NatsAMD64 != "" {
		imageList = append(imageList, images.NatsAMD64)
	}
	if images.NatsARM64 != "" {
		imageList = append(imageList, images.NatsARM64)
	}
	if images.NatsRISCV64 != "" {
		imageList = append(imageList, images.NatsRISCV64)
	}
	if images.NatsARM != "" {
		imageList = append(imageList, images.NatsARM)
	}

	controllers := remoteControlPlane.GetControllers()
	ctx := context.Background()
	for _, baseController := range controllers {
		controller, ok := baseController.(*rsc.RemoteController)
		if !ok {
			return util.NewInternalError("Could not convert Controller to Remote Controller")
		}
		platform, err := deployairgap.ResolvePlatform(controller.SystemAgent.AgentConfiguration.Arch)
		if err != nil {
			return fmt.Errorf("controller %s: %w", controller.Name, err)
		}
		engine, err := deployairgap.ResolveContainerEngine(controller.SystemAgent.AgentConfiguration.ContainerEngine)
		if err != nil {
			return fmt.Errorf("controller %s: %w", controller.Name, err)
		}
		if err := deployairgap.TransferAirgapImages(ctx, exe.ns.Name, controller.Host, &controller.SSH, platform, engine, imageList); err != nil {
			return fmt.Errorf("failed to transfer images to controller %s: %w", controller.Name, err)
		}
	}

	return nil
}

// transferSystemAgentImages transfers agent, router, and debugger images for system agents in airgap deployment
func (exe remoteControlPlaneExecutor) transferSystemAgentImages() error {
	remoteControlPlane, ok := exe.controlPlane.(*rsc.RemoteControlPlane)
	if !ok {
		return util.NewInternalError("Could not convert ControlPlane to Remote ControlPlane")
	}

	// Determine if this is initial deployment
	isInitial, err := deployairgap.IsInitialDeployment(exe.ns.Name)
	if err != nil {
		return fmt.Errorf("failed to determine deployment type: %w", err)
	}

	controllers := remoteControlPlane.GetControllers()
	for _, baseController := range controllers {
		controller, ok := baseController.(*rsc.RemoteController)
		if !ok {
			return util.NewInternalError("Could not convert Controller to Remote Controller")
		}

		// Skip if no system agent config
		if controller.SystemAgent == nil || controller.SystemAgent.AgentConfiguration == nil {
			continue
		}

		// Validate airgap requirements for system agent
		if err := deployairgap.ValidateAirgapRequirements(controller.SystemAgent.AgentConfiguration); err != nil {
			return fmt.Errorf("system agent for controller %s: %w", controller.Name, err)
		}

		// Resolve platform and container engine
		platform, err := deployairgap.ResolvePlatform(controller.SystemAgent.AgentConfiguration.Arch)
		if err != nil {
			return fmt.Errorf("system agent for controller %s: %w", controller.Name, err)
		}

		engine, err := deployairgap.ResolveContainerEngine(controller.SystemAgent.AgentConfiguration.ContainerEngine)
		if err != nil {
			return fmt.Errorf("system agent for controller %s: %w", controller.Name, err)
		}

		// Create a temporary RemoteAgent for image collection
		tempAgent := &rsc.RemoteAgent{
			Name:    controller.Name,
			Host:    controller.Host,
			SSH:     controller.SSH,
			Package: controller.SystemAgent.Package,
			Config:  controller.SystemAgent.AgentConfiguration,
		}

		// Collect required images
		images, err := deployairgap.CollectAgentImages(exe.ns.Name, tempAgent, remoteControlPlane, isInitial)
		if err != nil {
			return fmt.Errorf("failed to collect agent images for system agent %s: %w", controller.Name, err)
		}

		// Get router image for the platform
		routerImage, err := deployairgap.GetImageForPlatform(images, platform)
		if err != nil {
			return fmt.Errorf("failed to get router image for platform %s: %w", platform, err)
		}

		// Prepare image list (agent, router for platform, NATS, debugger if available)
		imageList := []string{images.Agent}
		if routerImage != "" {
			imageList = append(imageList, routerImage)
		}
		if images.NatsAMD64 != "" {
			imageList = append(imageList, images.NatsAMD64)
		}
		if images.NatsARM64 != "" {
			imageList = append(imageList, images.NatsARM64)
		}
		if images.NatsRISCV64 != "" {
			imageList = append(imageList, images.NatsRISCV64)
		}
		if images.NatsARM != "" {
			imageList = append(imageList, images.NatsARM)
		}
		if images.DebuggerAMD64 != "" {
			imageList = append(imageList, images.DebuggerAMD64)
		}
		if images.DebuggerARM64 != "" {
			imageList = append(imageList, images.DebuggerARM64)
		}
		if images.DebuggerRISCV64 != "" {
			imageList = append(imageList, images.DebuggerRISCV64)
		}
		if images.DebuggerARM != "" {
			imageList = append(imageList, images.DebuggerARM)
		}

		// Transfer images
		ctx := context.Background()
		if err := deployairgap.TransferAirgapImages(ctx, exe.ns.Name, controller.Host, &controller.SSH, platform, engine, imageList); err != nil {
			return fmt.Errorf("failed to transfer images to system agent %s: %w", controller.Name, err)
		}
	}

	return nil
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	// Check the namespace exists
	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return
	}

	// Read the input file
	controlPlane, err := rsc.UnmarshallRemoteControlPlane(opt.Yaml)
	if err != nil {
		return
	}

	// Validate control plane for multiple controllers
	if err := validateMultiControllerConfig(&controlPlane); err != nil {
		return nil, err
	}

	// Create exe Controllers
	controllers := controlPlane.GetControllers()
	controllerExecutors := make([]execute.Executor, len(controllers))
	for idx := range controllers {
		controller, ok := controllers[idx].(*rsc.RemoteController)
		if !ok {
			return nil, util.NewError("Could not convert Controller to Remote Controller")
		}
		exe, err := deployremotecontroller.NewExecutorWithoutParsing(opt.Namespace, &controlPlane, controller)
		if err != nil {
			return nil, err
		}
		controllerExecutors[idx] = exe
	}

	return newControlPlaneExecutor(controllerExecutors, ns, opt.Name, &controlPlane), nil
}

func runExecutors(executors []execute.Executor) error {
	if errs, _ := execute.ForParallel(executors); len(errs) > 0 {
		return execute.CoalesceErrors(errs)
	}
	return nil
}
