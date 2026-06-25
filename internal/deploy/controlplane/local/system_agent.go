package deploylocalcontrolplane

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

func buildLocalSystemAgentConfig(cp *rsc.LocalControlPlane, name string) (rsc.AgentConfiguration, error) {
	sys := cp.SystemAgent
	if sys == nil {
		return rsc.AgentConfiguration{}, util.NewInputError("Local Control Plane systemAgent is required")
	}

	var deployAgentConfig rsc.AgentConfiguration
	if sys.AgentConfiguration != nil {
		deployAgentConfig = *sys.AgentConfiguration
	}

	deploymentType := deploymentTypeNative
	if deployAgentConfig.DeploymentType != nil {
		deploymentType = *deployAgentConfig.DeploymentType
	} else if sys.Package.Container.Image != "" {
		deploymentType = deploymentTypeContainer
	}

	if deployAgentConfig.Host == nil || strings.TrimSpace(*deployAgentConfig.Host) == "" {
		ip, err := util.DetectLocalHostIPv4()
		if err != nil {
			return rsc.AgentConfiguration{}, util.NewInputError("Local Control Plane systemAgent requires spec.config.host or a detectable local IPv4 address")
		}
		deployAgentConfig.Host = &ip
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
		deployAgentConfig.Name = name
	}

	return deployAgentConfig, nil
}

func deployLocalSystemAgent(namespace string, cp *rsc.LocalControlPlane, name string, edgelet *install.LocalEdgelet) error {
	install.Verbose("Deploying system agent for local control plane " + name)

	deployAgentConfig, err := buildLocalSystemAgentConfig(cp, name)
	if err != nil {
		return err
	}

	configExe := deployagentconfig.NewRemoteExecutor(name, &deployAgentConfig, namespace, nil)
	if err := configExe.Execute(); err != nil {
		return fmt.Errorf("failed to deploy system agent configuration: %w", err)
	}

	user := install.IofogUser(cp.GetUser())
	user.Password = cp.GetUser().GetRawPassword()

	endpoint, err := cp.GetEndpoint()
	if err != nil {
		return err
	}

	opt, err := clientutil.ControllerClientOptions(context.Background(), namespace, endpoint)
	if err != nil {
		return err
	}

	if _, err := edgelet.Configure(endpoint, user, opt); err != nil {
		return fmt.Errorf("failed to provision system agent: %w", err)
	}
	return persistLocalSystemAgent(namespace, cp, name, deployAgentConfig, endpoint, configExe.GetAgentUUID())
}

func persistLocalSystemAgent(namespace string, cp *rsc.LocalControlPlane, name string, deployAgentConfig rsc.AgentConfiguration, endpoint, uuid string) error {
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return err
	}

	host := deployAgentConfig.Host
	if host == nil || strings.TrimSpace(*host) == "" {
		var agentInfo *client.AgentInfo
		err = clientutil.ExecuteWithAuthRetry(namespace, func(clt *client.Client) error {
			var err error
			agentInfo, err = clt.GetAgentByName(name)
			return err
		})
		if err != nil {
			return fmt.Errorf("failed to load system agent from controller: %w", err)
		}
		host = &agentInfo.Host
	}

	configCopy := deployAgentConfig
	agent := &rsc.LocalAgent{
		Name:               name,
		UUID:               uuid,
		Created:            util.NowUTC(),
		Host:               *host,
		ControllerEndpoint: endpoint,
		Airgap:             cp.Airgap,
		Config:             &configCopy,
	}
	if cp.SystemAgent != nil {
		agent.Package = cp.SystemAgent.Package
		agent.Scripts = cp.SystemAgent.Scripts
	}

	if err := ns.UpdateAgent(agent); err != nil {
		return err
	}
	return config.Flush()
}
