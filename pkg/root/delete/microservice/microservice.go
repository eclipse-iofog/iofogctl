package deletemicroservice

import (
	"fmt"
	"github.com/eclipse-iofog/cli/pkg/config"
)

type microservice struct {
	configManager *config.Manager
}

func new() *microservice {
	c := &microservice{}
	c.configManager = config.NewManager()
	return c
}

func (ctrl *microservice) execute(namespace, name string) error {
	// TODO (Serge) Execute back-end logic

	// Update configuration
	err := ctrl.configManager.DeleteMicroservice(namespace, name)

	// TODO (Serge) Handle config file error, retry..?

	if err == nil {
		fmt.Printf("\nMicroservice %s/%s successfully deleted.\n", namespace, name)
	}
	return err
}
