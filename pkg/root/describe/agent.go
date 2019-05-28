package describe

import (
	"github.com/eclipse-iofog/cli/pkg/config"
)

type agentExecutor struct {
	configManager *config.Manager
}

func newAgentExecutor() *agentExecutor {
	a := &agentExecutor{}
	a.configManager = config.NewManager()
	return a
}

func (ns *agentExecutor) execute(namespace string, name string) error {
	agent, err := ns.configManager.GetAgent(namespace, name)
	if err != nil {
		return err
	}
	if err = print(agent); err != nil {
		return err
	}
	return nil
}
