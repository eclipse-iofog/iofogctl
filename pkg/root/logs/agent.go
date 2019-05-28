package logs

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
	return nil
}
