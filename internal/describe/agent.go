package describe

import (
	"github.com/eclipse-iofog/cli/internal/config"
)

type agentExecutor struct {
	namespace string
	name      string
}

func newAgentExecutor(namespace, name string) *agentExecutor {
	a := &agentExecutor{}
	a.namespace = namespace
	a.name = name
	return a
}

func (exe *agentExecutor) Execute() error {
	agent, err := config.GetAgent(exe.namespace, exe.name)
	if err != nil {
		return err
	}
	if err = print(agent); err != nil {
		return err
	}
	return nil
}
