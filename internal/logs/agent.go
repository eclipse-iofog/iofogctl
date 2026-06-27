package logs

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type agentExecutor struct {
	namespace string
	name      string
	logConfig *LogTailConfig
}

func newAgentExecutor(namespace, name string, logConfig *LogTailConfig) *agentExecutor {
	exe := &agentExecutor{}
	exe.namespace = namespace
	exe.name = name
	exe.logConfig = logConfig
	return exe
}

func (exe *agentExecutor) GetName() string {
	return exe.name
}

func (exe *agentExecutor) Execute() error {
	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	// Update local cache based on Controller
	if err := clientutil.SyncAgentInfo(exe.namespace); err != nil {
		return err
	}

	// Get agent config
	baseAgent, err := ns.GetAgent(exe.name)
	if err != nil {
		return err
	}

	switch agent := baseAgent.(type) {
	case *rsc.LocalAgent:
		sdkCfg := localAgentSDKConfig(agent)
		lc, err := install.NewLocalContainerClient(install.LocalContainerEngineForHostOps(sdkCfg), sdkCfg)
		if err != nil {
			return err
		}
		stdout, stderr, err := lc.GetLogsByName(install.EdgeletContainerName)
		if err != nil {
			return err
		}

		printContainerLogs(stdout, stderr)

		return nil
	case *rsc.RemoteAgent:
		util.SpinStart("Connecting to Agent Logs")

		clt, err := clientutil.NewControllerClient(exe.namespace)
		if err != nil {
			util.SpinHandlePromptComplete()
			return err
		}

		agentInfo, err := clt.GetAgentByName(exe.name)
		if err != nil {
			util.SpinHandlePromptComplete()
			return fmt.Errorf("failed to get Agent by name: %s", err.Error())
		}

		opts := exe.logConfig.ToSDKOptions()
		return runRemoteLogStream(clt, func(clt *client.Client) (*client.LogSession, error) {
			return clt.DialFogLogs(agentInfo.UUID, opts)
		})
	}

	return nil
}

func localAgentSDKConfig(agent *rsc.LocalAgent) *client.AgentConfiguration {
	if agent == nil || agent.Config == nil {
		return nil
	}
	return &agent.Config.AgentConfiguration
}
