package deletecontroller

import (
	"github.com/eclipse-iofog/cli/pkg/config"
	"fmt"
)

type controller struct {
	configManager *config.Manager
}

func new() *controller {
	c := &controller{}
	c.configManager = config.NewManager()
	return c
}

func (ctrl *controller) execute(namespace, name string) error {
	// TODO (Serge) Execute back-end logic

	// Update configuration
	err := ctrl.configManager.DeleteController(namespace, name)

	// TODO (Serge) Handle config file error, retry..?

	if err == nil {
		fmt.Printf("\nController %s/%s successfully deleted.\n", namespace, name)
	}
	return err
}