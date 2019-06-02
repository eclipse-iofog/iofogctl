package get

import (
	"github.com/eclipse-iofog/cli/internal/config"
)

type agentExecutor struct {
	namespace string
}

func newAgentExecutor(namespace string) *agentExecutor {
	a := &agentExecutor{}
	a.namespace = namespace
	return a
}

func (exe *agentExecutor) Execute() error {
	agents, err := config.GetAgents(exe.namespace)
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
