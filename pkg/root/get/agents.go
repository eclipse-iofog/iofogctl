package get

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

func (agent *agentExecutor) execute(namespace string) error {
	agents, err := agent.configManager.GetAgents(namespace)
	if err != nil {
		return err
	}
	rows := make([]row, len(agents))
	for idx, agent := range agents {
		rows[idx].name = agent.Name
		// TODO: (Serge) Get runtime info
		rows[idx].status = "-"
		rows[idx].age = "-"
	}
	err = print(rows)
	return err
}
