package deployagent

import (
	"fmt"
	"regexp"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type localExecutor struct {
	isSystem         bool
	namespace        string
	agent            *rsc.LocalAgent
	client           *install.LocalContainer
	localAgentConfig *install.LocalAgentConfig
}

func newLocalExecutor(namespace string, agent *rsc.LocalAgent, isSystem bool) (*localExecutor, error) {
	client, err := install.NewLocalContainerClient()
	if err != nil {
		return nil, err
	}
	if agent.Config == nil {
		agent.Config = &rsc.AgentConfiguration{}
	}
	// Get Controller LocalContainerConfig
	controllerContainerConfig := install.NewLocalControllerConfig("", install.Credentials{}, install.Auth{}, install.Database{}, install.Events{}, nil)
	return &localExecutor{
		isSystem:  isSystem,
		namespace: namespace,
		agent:     agent,
		client:    client,
		localAgentConfig: install.NewLocalAgentConfig(
			agent.Name,
			agent.Container.Image,
			controllerContainerConfig,
			install.Credentials{
				User:     agent.Container.Credentials.User,
				Password: agent.Container.Credentials.Password,
			},
			isSystem,
			agent.Config.TimeZone,
		),
	}, nil
}

func (exe *localExecutor) ProvisionAgent() (string, error) {
	// Get agent
	agent := install.NewLocalAgent(exe.localAgentConfig, exe.client)

	// Get user
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
			return "", err
		}
	}

	// Configure the agent with Controller details
	user := install.IofogUser(controlPlane.GetUser())
	user.Password = controlPlane.GetUser().GetRawPassword()
	return agent.Configure(controllerEndpoint, user)
}

func (exe *localExecutor) GetName() string {
	return exe.agent.Name
}

func (exe *localExecutor) Execute() error {
	// Deploy agent image
	util.SpinStart("Deploying Agent container")
	if exe.agent.Container.Image == "" {
		exe.agent.Container.Image = exe.localAgentConfig.DefaultImage
	}

	// If container already exists, clean it
	agentContainerName := exe.localAgentConfig.ContainerName
	if _, err := exe.client.GetContainerByName(agentContainerName); err == nil {
		if err := exe.client.CleanContainer(agentContainerName); err != nil {
			return err
		}
	}

	if _, err := exe.client.DeployContainer(&exe.localAgentConfig.LocalContainerConfig); err != nil {
		return err
	}

	// Wait for agent
	util.SpinStart("Waiting for Agent")
	if err := exe.client.WaitForCommand(
		install.GetLocalContainerName("agent", exe.isSystem),
		regexp.MustCompile("ioFog daemon[ |\t]*: RUNNING"),
		"iofog-agent",
		"status",
	); err != nil {
		if cleanErr := exe.client.CleanContainer(agentContainerName); cleanErr != nil {
			util.PrintNotify(fmt.Sprintf("Could not clean container: %v", agentContainerName))
		}
		return err
	}

	// Provision agent
	util.SpinStart("Provisioning Agent")
	uuid, err := exe.ProvisionAgent()
	if err != nil {
		if cleanErr := exe.client.CleanContainer(agentContainerName); cleanErr != nil {
			util.PrintNotify(fmt.Sprintf("Could not clean container: %v", agentContainerName))
		}
		return err
	}

	// Return new Agent config because variable is a pointer
	exe.agent.Host = fmt.Sprintf("%s:%s", exe.localAgentConfig.Host, exe.localAgentConfig.Ports[0].Host)
	exe.agent.UUID = uuid

	return nil
}
