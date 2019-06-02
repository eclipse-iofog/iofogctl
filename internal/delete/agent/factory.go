package deleteagent

import (
	"github.com/eclipse-iofog/cli/internal/config"
)

type Executor interface {
	Execute() error
}

func NewExecutor(namespace, name string) (Executor, error) {
	agent, err := config.GetAgent(namespace, name)
	if err != nil {
		return nil, err
	}

	// Local executor
	if agent.Host == "localhost" {
		return newLocalExecutor(namespace, name), nil
	}

	// Default executor
	return newRemoteExecutor(namespace, name), nil
}
