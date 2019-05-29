package deleteagent

import (
	"fmt"
	"github.com/eclipse-iofog/cli/internal/config"
)

type localExecutor struct {
	configManager *config.Manager
	namespace     string
	agent         config.Agent
}

func newLocalExecutor(cfg *config.Manager, ns string, ctrl config.Agent) *localExecutor {
	exe := &localExecutor{}
	exe.configManager = cfg
	exe.namespace = ns
	exe.agent = ctrl
	return exe
}

func (exe *localExecutor) Execute() error {
	// TODO (Serge) Execute back-end logic

	// Update configuration
	err := exe.configManager.DeleteAgent(exe.namespace, exe.agent.Name)

	// TODO (Serge) Handle config file error, retry..?

	if err == nil {
		fmt.Printf("\nAgent %s/%s successfully deleted.\n", exe.namespace, exe.agent.Name)
	}
	return err
}
