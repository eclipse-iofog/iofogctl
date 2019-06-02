package deletemicroservice

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
	err := config.DeleteMicroservice(namespace, name)

	// TODO (Serge) Handle config file error, retry..?

	if err == nil {
		fmt.Printf("\nMicroservice %s/%s successfully deleted.\n", namespace, name)
	}
	return err
}
