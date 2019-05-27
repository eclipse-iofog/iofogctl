package logs

import (
	"github.com/eclipse-iofog/cli/pkg/config"
)

type controllerExecutor struct {
	configManager *config.Manager
}

func newControllerExecutor() *controllerExecutor {
	c := &controllerExecutor{}
	c.configManager = config.NewManager()
	return c
}

func (ns *controllerExecutor) execute(namespace string, name string) error {
	return nil
}