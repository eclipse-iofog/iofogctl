package deploymicroservice

import (
	"github.com/eclipse-iofog/cli/pkg/config"
	"fmt"
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
	configEntry := config.Microservice{ Name: name }
	err := ctrl.configManager.AddMicroservice(namespace, configEntry)

	// TODO (Serge) Handle config file error, retry..?

	if err == nil {
		fmt.Printf("\nMicroservice %s/%s successfully deployed.\n", namespace, name)
	}
	return err
}