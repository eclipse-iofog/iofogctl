package deploymicroservice

import (
	"fmt"
	"github.com/eclipse-iofog/cli/internal/config"
)

type microservice struct {
}

func New() *microservice {
	c := &microservice{}
	return c
}

func (ctrl *microservice) Execute(namespace, name string) error {
	// TODO (Serge) Execute back-end logic

	// Update configuration
	configEntry := config.Microservice{Name: name}
	err := config.AddMicroservice(namespace, configEntry)

	// TODO (Serge) Handle config file error, retry..?

	if err == nil {
		fmt.Printf("\nMicroservice %s/%s successfully deployed.\n", namespace, name)
	}
	return err
}
