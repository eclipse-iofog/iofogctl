package deployagent

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
	configEntry := config.Agent{Name: name, User: "none"}
	err := ctrl.configManager.AddAgent(namespace, configEntry)

	// TODO (Serge) Handle config file error, retry..?

	if err == nil {
		fmt.Printf("\nAgent %s/%s successfully deployed.\n", namespace, name)
	}
	return err
}
