package deployremotecontrolplane

import (
	"context"
	"fmt"
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	deployagentconfig "github.com/eclipse-iofog/iofogctl/internal/deploy/agentconfig"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	iutil "github.com/eclipse-iofog/iofogctl/internal/util"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

const (
	deploymentTypeNative    = "native"
	deploymentTypeContainer = "container"
)

func resolveSystemAgentDeploymentType(systemAgent *rsc.SystemAgentConfig) string {
	if systemAgent != nil && systemAgent.AgentConfiguration != nil && systemAgent.AgentConfiguration.DeploymentType != nil {
		return *systemAgent.AgentConfiguration.DeploymentType
	}
	if systemAgent != nil && systemAgent.Package.Container.Image != "" {
		return deploymentTypeContainer
	}
	return deploymentTypeNative
}

func buildFirstSystemAgentConfig(cp *rsc.RemoteControlPlane, ctrl *rsc.RemoteController, systemAgent *rsc.SystemAgentConfig) rsc.AgentConfiguration {
	deploymentType := resolveSystemAgentDeploymentType(systemAgent)

	var deployAgentConfig rsc.AgentConfiguration
	if systemAgent != nil && systemAgent.AgentConfiguration != nil {
		deployAgentConfig = *systemAgent.AgentConfiguration
		if deployAgentConfig.Host == nil {
			deployAgentConfig.Host = defaultSystemAgentHost(cp, ctrl)
		}
		if deployAgentConfig.DeploymentType == nil {
			deployAgentConfig.DeploymentType = iutil.MakeStrPtr(deploymentType)
		}
	} else {
		host := defaultSystemAgentHost(cp, ctrl)
		upstreamRouters := []string{}
		upstreamNatsServers := []string{}
		deployAgentConfig = rsc.AgentConfiguration{
			Name: ctrl.Name,
			Arch: iutil.MakeStrPtr("auto"),
			AgentConfiguration: client.AgentConfiguration{
				IsSystem:            iutil.MakeBoolPtr(true),
				DeploymentType:      iutil.MakeStrPtr(deploymentType),
				Host:                host,
				RouterConfig:        deployagentconfig.DefaultRouterConfig(),
				UpstreamRouters:     &upstreamRouters,
				UpstreamNatsServers: &upstreamNatsServers,
			},
		}
	}

	deployAgentConfig.IsSystem = iutil.MakeBoolPtr(true)
	if deployAgentConfig.DeploymentType == nil {
		deployAgentConfig.DeploymentType = iutil.MakeStrPtr(deploymentType)
	}
	if deployAgentConfig.UpstreamRouters == nil {
		emptyRouters := []string{}
		deployAgentConfig.UpstreamRouters = &emptyRouters
	}
	if deployAgentConfig.UpstreamNatsServers == nil {
		emptyNats := []string{}
		deployAgentConfig.UpstreamNatsServers = &emptyNats
	}

	deployagentconfig.ApplySystemAgentDefaults(&deployAgentConfig)

	if deployAgentConfig.Name == "" {
		deployAgentConfig.Name = ctrl.Name
	}
	return deployAgentConfig
}

func buildNextSystemAgentConfig(cp *rsc.RemoteControlPlane, ctrl *rsc.RemoteController, systemAgent *rsc.SystemAgentConfig) rsc.AgentConfiguration {
	deploymentType := resolveSystemAgentDeploymentType(systemAgent)

	var deployAgentConfig rsc.AgentConfiguration
	if systemAgent != nil && systemAgent.AgentConfiguration != nil {
		deployAgentConfig = *systemAgent.AgentConfiguration
		if deployAgentConfig.Host == nil {
			deployAgentConfig.Host = defaultSystemAgentHost(cp, ctrl)
		}
		if deployAgentConfig.DeploymentType == nil {
			deployAgentConfig.DeploymentType = iutil.MakeStrPtr(deploymentType)
		}
		if deployAgentConfig.UpstreamRouters == nil {
			upstreamRouters := []string{"default-router"}
			deployAgentConfig.UpstreamRouters = &upstreamRouters
		} else if !containsString(*deployAgentConfig.UpstreamRouters, "default-router") {
			*deployAgentConfig.UpstreamRouters = append(*deployAgentConfig.UpstreamRouters, "default-router")
		}
		if deployAgentConfig.UpstreamNatsServers == nil {
			upstreamNatsServers := []string{"default-nats-hub"}
			deployAgentConfig.UpstreamNatsServers = &upstreamNatsServers
		} else if !containsString(*deployAgentConfig.UpstreamNatsServers, "default-nats-hub") {
			*deployAgentConfig.UpstreamNatsServers = append(*deployAgentConfig.UpstreamNatsServers, "default-nats-hub")
		}
	} else {
		host := defaultSystemAgentHost(cp, ctrl)
		upstreamRouters := []string{"default-router"}
		upstreamNatsServers := []string{"default-nats-hub"}
		deployAgentConfig = rsc.AgentConfiguration{
			Name: ctrl.Name,
			Arch: iutil.MakeStrPtr("auto"),
			AgentConfiguration: client.AgentConfiguration{
				IsSystem:            iutil.MakeBoolPtr(true),
				DeploymentType:      iutil.MakeStrPtr(deploymentType),
				Host:                host,
				RouterConfig:        deployagentconfig.DefaultRouterConfig(),
				UpstreamRouters:     &upstreamRouters,
				UpstreamNatsServers: &upstreamNatsServers,
			},
		}
	}

	deployAgentConfig.IsSystem = iutil.MakeBoolPtr(true)
	if deployAgentConfig.DeploymentType == nil {
		deployAgentConfig.DeploymentType = iutil.MakeStrPtr(deploymentType)
	}

	deployagentconfig.ApplySystemAgentDefaults(&deployAgentConfig)

	if deployAgentConfig.Name == "" {
		deployAgentConfig.Name = ctrl.Name
	}
	return deployAgentConfig
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func defaultSystemAgentHost(cp *rsc.RemoteControlPlane, ctrl *rsc.RemoteController) *string {
	host := resolveControllerAPIHost(cp, ctrl)
	if host == "" {
		return nil
	}
	return &host
}

func deployRemoteSystemAgent(namespace string, cp *rsc.RemoteControlPlane, ctrl *rsc.RemoteController, deployAgentConfig rsc.AgentConfiguration) error {
	configExe := deployagentconfig.NewRemoteExecutor(ctrl.Name, &deployAgentConfig, namespace, nil)
	if err := configExe.Execute(); err != nil {
		return fmt.Errorf("failed to deploy system agent configuration: %w", err)
	}

	agentUUID := configExe.GetAgentUUID()
	edgelet, err := BuildRemoteEdgelet(cp, ctrl, agentUUID)
	if err != nil {
		return err
	}

	endpoint, err := ResolveControllerHostEndpoint(cp, ctrl)
	if err != nil {
		return err
	}

	user := install.IofogUser(cp.GetUser())
	user.Password = cp.GetUser().GetRawPassword()
	opt, err := clientutil.ControllerClientOptions(context.Background(), namespace, endpoint)
	if err != nil {
		return err
	}
	if _, err := edgelet.Configure(endpoint, user, opt); err != nil {
		return fmt.Errorf("failed to provision system agent: %w", err)
	}

	return persistRemoteSystemAgent(namespace, cp, ctrl, deployAgentConfig, endpoint, agentUUID)
}

func persistRemoteSystemAgent(namespace string, cp *rsc.RemoteControlPlane, ctrl *rsc.RemoteController, deployAgentConfig rsc.AgentConfiguration, endpoint, uuid string) error {
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return err
	}

	host := ctrl.Host
	if deployAgentConfig.Host != nil && strings.TrimSpace(*deployAgentConfig.Host) != "" {
		host = *deployAgentConfig.Host
	}

	configCopy := deployAgentConfig
	agent := &rsc.RemoteAgent{
		Name:               ctrl.Name,
		UUID:               uuid,
		Created:            util.NowUTC(),
		Host:               host,
		SSH:                ctrl.SSH,
		ControllerEndpoint: endpoint,
		Airgap:             cp.Airgap,
		Config:             &configCopy,
	}
	if ctrl.SystemAgent != nil {
		agent.Package = ctrl.SystemAgent.Package
		agent.Scripts = ctrl.SystemAgent.Scripts
	}

	if err := ns.UpdateAgent(agent); err != nil {
		return err
	}
	return config.Flush()
}

func deploySystemAgent(namespace string, cp *rsc.RemoteControlPlane, ctrl *rsc.RemoteController) error {
	install.Verbose("Deploying system agent for controller " + ctrl.Name)
	cfg := buildFirstSystemAgentConfig(cp, ctrl, ctrl.SystemAgent)
	return deployRemoteSystemAgent(namespace, cp, ctrl, cfg)
}

func DeployNextSystemAgent(namespace string, cp *rsc.RemoteControlPlane, ctrl *rsc.RemoteController) error {
	install.Verbose("Deploying next-system agent for controller " + ctrl.Name)
	cfg := buildNextSystemAgentConfig(cp, ctrl, ctrl.SystemAgent)
	return deployRemoteSystemAgent(namespace, cp, ctrl, cfg)
}

func deployRemoteSystemAgents(namespace string, cp *rsc.RemoteControlPlane) error {
	controllers := cp.GetControllers()
	for idx, baseController := range controllers {
		controller, ok := baseController.(*rsc.RemoteController)
		if !ok {
			return util.NewInternalError("Could not convert Controller to Remote Controller")
		}
		if idx == 0 {
			if err := deploySystemAgent(namespace, cp, controller); err != nil {
				return fmt.Errorf("failed to deploy system agent for first controller: %w", err)
			}
			continue
		}
		if err := DeployNextSystemAgent(namespace, cp, controller); err != nil {
			return fmt.Errorf("failed to deploy next-system agent for controller %d: %w", idx, err)
		}
	}
	return nil
}
