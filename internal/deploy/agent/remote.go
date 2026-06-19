package deployagent

import (
	"context"
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	deployairgap "github.com/eclipse-iofog/iofogctl/internal/deploy/airgap"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"

	// clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	iutil "github.com/eclipse-iofog/iofogctl/internal/util"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type remoteExecutor struct {
	namespace string
	agent     *rsc.RemoteAgent
}

func newRemoteExecutor(namespace string, agent *rsc.RemoteAgent) *remoteExecutor {
	exe := &remoteExecutor{}
	exe.namespace = namespace
	exe.agent = agent

	return exe
}

func (exe *remoteExecutor) GetName() string {
	return exe.agent.Name
}

func (exe *remoteExecutor) ProvisionAgent() (string, error) {
	var agent *install.RemoteAgent
	var err error
	// If DeploymentType is nil, default to "container"
	// Use NewRemoteContainerAgent if DeploymentType is nil or "container"
	// Check if Config is nil and initialize if needed
	if exe.agent.Config == nil {
		exe.agent.Config = &rsc.AgentConfiguration{}
	}

	// Check DeploymentType (Config is guaranteed to be non-nil now)
	if exe.agent.Config.DeploymentType == nil || *exe.agent.Config.DeploymentType == "container" {
		// Use NewRemoteContainerAgent
		agent, err = install.NewRemoteContainerAgent(
			exe.agent.SSH.User,
			exe.agent.Host,
			exe.agent.SSH.Port,
			exe.agent.SSH.KeyFile,
			exe.agent.Name,
			exe.agent.UUID,
			exe.agent.Config.TimeZone,
		)
		if err == nil {
			// Set airgap flag if enabled
			agent.SetAirgap(exe.agent.Airgap)
		}
	} else {
		// Use NewRemoteAgent for "native" deployment type
		agent, err = install.NewRemoteAgent(
			exe.agent.SSH.User,
			exe.agent.Host,
			exe.agent.SSH.Port,
			exe.agent.SSH.KeyFile,
			exe.agent.Name,
			exe.agent.UUID,
		)
	}

	if err != nil {
		return "", err
	}

	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return "", err
	}
	controlPlane, err := ns.GetControlPlane()
	if err != nil {
		return "", err
	}
	// Try Agent-specific endpoint first
	controllerEndpoint := exe.agent.GetControllerEndpoint()
	if controllerEndpoint == "" {
		controllerEndpoint, err = controlPlane.GetEndpoint()
		if err != nil {
			return "", util.NewError("Failed to retrieve Controller endpoint!")
		}
	}

	// Configure the agent with Controller details
	user := install.IofogUser(controlPlane.GetUser())
	user.Password = controlPlane.GetUser().GetRawPassword()
	// Get Config before provision and set iofog-agent config
	agentConfig := exe.agent.GetConfig()

	// Check if agentConfig is empty
	if agentConfig == nil {
		util.PrintNotify(fmt.Sprintf("Skipping initial agent configuration for %s as agent config parameters are empty. Default config parameters will be used.", exe.agent.Name))

	} else {
		var fogType *string
		if agentConfig.FogType == nil {
			auto := "auto"
			fogType = &auto
		} else {
			fogType = agentConfig.FogType
		}
		err = agent.SetInitialConfig(
			agentConfig.Name,
			// agentConfig.Location,
			// agentConfig.Latitude,
			// agentConfig.Longitude,
			// agentConfig.Description,
			*fogType,
			agentConfig.AgentConfiguration, // Pass the embedded client.AgentConfiguration
		)
		if err != nil {
			return "", err
		}
	}
	return agent.Configure(controllerEndpoint, user)
}

// Deploy iofog-agent stack on an agent host
func (exe *remoteExecutor) Execute() (err error) {
	// Get Control Plane
	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	controlPlane, err := ns.GetControlPlane()
	if err != nil || len(controlPlane.GetControllers()) == 0 {
		util.PrintError("You must deploy a Controller to a namespace before deploying any Agents")
		return
	}

	var agent *install.RemoteAgent
	// If DeploymentType is nil, default to "container"
	// Use NewRemoteContainerAgent if DeploymentType is nil, "container", or if container image is specified
	// Check if Config is nil and initialize if needed
	if exe.agent.Config == nil {
		exe.agent.Config = &rsc.AgentConfiguration{}
	}

	useContainer := false
	// Check DeploymentType (Config is guaranteed to be non-nil now)
	if exe.agent.Config.DeploymentType == nil {
		useContainer = true
	} else if *exe.agent.Config.DeploymentType == "container" {
		useContainer = true
	}
	// Check if container image is specified (Package is a value type, always initialized)
	if !useContainer && exe.agent.Package.Container.Image != "" {
		useContainer = true
	}

	if useContainer {
		exe.agent.Config.DeploymentType = iutil.MakeStrPtr("container")
		// Use NewRemoteContainerAgent
		agent, err = install.NewRemoteContainerAgent(
			exe.agent.SSH.User,
			exe.agent.Host,
			exe.agent.SSH.Port,
			exe.agent.SSH.KeyFile,
			exe.agent.Name,
			exe.agent.UUID,
			exe.agent.Config.TimeZone,
		)
	} else {
		// Use NewRemoteAgent for "native" deployment type
		exe.agent.Config.DeploymentType = iutil.MakeStrPtr("native")
		agent, err = install.NewRemoteAgent(
			exe.agent.SSH.User,
			exe.agent.Host,
			exe.agent.SSH.Port,
			exe.agent.SSH.KeyFile,
			exe.agent.Name,
			exe.agent.UUID,
		)
	}
	if err != nil {
		return err
	}

	// Set custom scripts
	if exe.agent.Scripts != nil {
		if err := agent.CustomizeProcedures(
			exe.agent.Scripts.Directory,
			&exe.agent.Scripts.AgentProcedures); err != nil {
			return err
		}
	}
	if exe.agent.Package.Container.Image != "" {
		// Set Image
		agent.SetContainerImage(exe.agent.Package.Container.Image)

	} else if exe.agent.Package.Version != "" {
		// Set version
		agent.SetVersion(exe.agent.Package.Version)
	}
	// // Set version
	// agent.SetVersion(exe.agent.Package.Version)
	// agent.SetRepository(exe.agent.Package.Repo, exe.agent.Package.Token)

	// Handle airgap deployment if enabled
	if exe.agent.Airgap {
		// Validate airgap requirements
		if err := deployairgap.ValidateAirgapRequirements(exe.agent.Config); err != nil {
			return fmt.Errorf("airgap deployment requires valid configuration: %w", err)
		}

		// Resolve platform and container engine
		platform, err := deployairgap.ResolvePlatform(exe.agent.Config.FogType)
		if err != nil {
			return fmt.Errorf("failed to resolve platform: %w", err)
		}

		engine, err := deployairgap.ResolveContainerEngine(exe.agent.Config.AgentConfiguration.ContainerEngine)
		if err != nil {
			return fmt.Errorf("failed to resolve container engine: %w", err)
		}

		// Determine if this is initial deployment
		isInitial, err := deployairgap.IsInitialDeployment(exe.namespace)
		if err != nil {
			return fmt.Errorf("failed to determine deployment type: %w", err)
		}

		// Collect required images
		// For Kubernetes control planes, pass nil and always fetch from catalog (existing control plane)
		var remoteControlPlane *rsc.RemoteControlPlane
		if remoteCP, ok := controlPlane.(*rsc.RemoteControlPlane); ok {
			remoteControlPlane = remoteCP
		} else {
			// For Kubernetes or other control planes, treat as existing control plane
			// and fetch images from catalog items
			isInitial = false
		}

		images, err := deployairgap.CollectAgentImages(exe.namespace, exe.agent, remoteControlPlane, isInitial)
		if err != nil {
			return fmt.Errorf("failed to collect agent images: %w", err)
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
		if images.Nats != "" {
			imageList = append(imageList, images.Nats)
		}
		if images.Debugger != "" {
			imageList = append(imageList, images.Debugger)
		}

		// Transfer images before bootstrap
		ctx := context.Background()
		if err := deployairgap.TransferAirgapImages(ctx, exe.namespace, exe.agent.Host, &exe.agent.SSH, platform, engine, imageList); err != nil {
			return fmt.Errorf("failed to transfer airgap images: %w", err)
		}
	}

	// Try the deploy
	err = agent.Bootstrap()
	if err != nil {
		return
	}

	uuid, err := exe.ProvisionAgent()
	if err != nil {
		return err
	}

	// Return the Agent through pointer
	exe.agent.UUID = uuid
	exe.agent.Created = util.NowUTC()
	return nil
}

func ValidateRemoteAgent(agent *rsc.RemoteAgent) error {
	if err := util.IsLowerAlphanumeric("Agent", agent.Name); err != nil {
		return err
	}
	if agent.Name == iofog.VanillaRouterAgentName {
		return util.NewInputError(fmt.Sprintf("%s is a reserved name and cannot be used for an Agent", iofog.VanillaRouterAgentName))
	}
	if (agent.Host != "localhost" && agent.Host != "127.0.0.1") && (agent.Host == "" || agent.SSH.User == "" || agent.SSH.KeyFile == "") {
		return util.NewInputError("For Agents you must specify non-empty values for host, user, and keyfile")
	}
	return nil
}
