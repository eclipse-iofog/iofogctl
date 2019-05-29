package logs

import (
	"github.com/eclipse-iofog/cli/internal/config"
)

type microserviceExecutor struct {
	configManager *config.Manager
}

func newMicroserviceExecutor() *microserviceExecutor {
	m := &microserviceExecutor{}
	m.configManager = config.NewManager()
	return m
}

func (ns *microserviceExecutor) Execute(namespace string, name string) error {
	return nil
}
