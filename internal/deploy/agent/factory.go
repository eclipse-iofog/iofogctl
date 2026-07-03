package deployagent

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	agentconfig "github.com/eclipse-iofog/iofogctl/internal/deploy/agentconfig"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type AgentDeployExecutor interface {
	execute.Executor
	GetHost() string
	GetTags() *[]string
	GetConfig() *rsc.AgentConfiguration
}

type facadeExecutor struct {
	isSystem  bool
	exe       execute.Executor
	agent     rsc.Agent
	namespace string
	tags      *[]string
}

func (facade *facadeExecutor) GetHost() string {
	return facade.agent.GetHost()
}

func (facade *facadeExecutor) GetTags() *[]string {
	return facade.tags
}

func (facade *facadeExecutor) GetConfig() *rsc.AgentConfiguration {
	return facade.agent.GetConfig()
}

func (facade *facadeExecutor) Execute() (err error) {
	// Check the namespace exists
	ns, err := config.GetNamespace(facade.namespace)
	if err != nil {
		return err
	}
	controlPlane, err := ns.GetControlPlane()
	if err != nil {
		return err
	}

	// Check Controller exists
	if len(controlPlane.GetControllers()) == 0 {
		return rsc.NewNoControlPlaneError(facade.namespace)
	}

	if !facade.isSystem || install.IsVerbose() {
		util.SpinStart(fmt.Sprintf("Deploying agent %s", facade.GetName()))
	}

	if err = facade.exe.Execute(); err != nil {
		return
	}

	if agentConfig := facade.agent.GetConfig(); agentConfig != nil {
		configExe := agentconfig.NewRemoteExecutor(facade.agent.GetName(), agentConfig, facade.namespace, facade.tags)
		configExe.SetHost(agentconfig.ResolveAgentAPIHost(facade.agent.GetHost(), agentConfig))
		if err := configExe.Execute(); err != nil {
			return err
		}
	}

	uuid, err := facade.ProvisionAgent()
	if err != nil {
		return err
	}
	facade.agent.SetUUID(uuid)
	facade.agent.SetCreatedTime(util.NowUTC())

	if err = ns.UpdateAgent(facade.agent); err != nil {
		return
	}

	return config.Flush()
}

func (facade *facadeExecutor) GetName() string {
	return facade.exe.GetName()
}

func (facade *facadeExecutor) ProvisionAgent() (string, error) {
	// Required for attach
	provisionExecutor, ok := facade.exe.(execute.ProvisioningExecutor)
	if !ok {
		return "", util.NewInternalError("Facade executor: Could not convert executor")
	}
	return provisionExecutor.ProvisionAgent()
}

func newFacadeExecutor(exe execute.Executor, namespace string, agent rsc.Agent, isSystem bool, tags *[]string) execute.Executor {
	return &facadeExecutor{
		exe:       exe,
		namespace: namespace,
		isSystem:  isSystem,
		agent:     agent,
		tags:      tags,
	}
}

func NewRemoteExecutor(namespace string, agent *rsc.RemoteAgent, isSystem bool) (execute.Executor, error) {
	if err := util.IsLowerAlphanumeric("Agent", agent.GetName()); err != nil {
		return nil, err
	}

	if err := agent.Sanitize(); err != nil {
		return nil, err
	}
	if err := agent.ValidateSSH(); err != nil {
		return nil, err
	}
	return newFacadeExecutor(newRemoteExecutor(namespace, agent), namespace, agent, isSystem, nil), nil
}

func NewLocalExecutor(namespace string, agent *rsc.LocalAgent, isSystem bool) (execute.Executor, error) {
	if err := util.IsLowerAlphanumeric("Agent", agent.GetName()); err != nil {
		return nil, err
	}

	exe, err := newLocalExecutor(namespace, agent, isSystem)
	if err != nil {
		return nil, err
	}
	return newFacadeExecutor(exe, namespace, agent, isSystem, nil), nil
}
