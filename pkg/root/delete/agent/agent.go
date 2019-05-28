package deleteagent

import (
	"fmt"
	"github.com/eclipse-iofog/cli/pkg/config"
)

type agent struct {
	configManager *config.Manager
}

func new() *agent {
	c := &agent{}
	c.configManager = config.NewManager()
	return c
}

func (ctrl *agent) execute(namespace, name string) error {
	// TODO (Serge) Execute back-end logic

	// Update configuration
	err := ctrl.configManager.DeleteAgent(namespace, name)

	// TODO (Serge) Handle config file error, retry..?

	if err == nil {
		fmt.Printf("\nAgent %s/%s successfully deleted.\n", namespace, name)
	}
	return err
}
