package logs

import (
	"github.com/eclipse-iofog/cli/pkg/config"
)

type microserviceExecutor struct {
	configManager *config.Manager
}

func newMicroserviceExecutor() *microserviceExecutor {
	m := &microserviceExecutor{}
	m.configManager = config.NewManager()
	return m
}

func (ns *microserviceExecutor) execute(namespace string, name string) error {
	return nil
}
