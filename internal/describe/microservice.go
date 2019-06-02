package describe

import (
	"github.com/eclipse-iofog/cli/internal/config"
)

type microserviceExecutor struct {
	namespace string
	name      string
}

func newMicroserviceExecutor(namespace, name string) *microserviceExecutor {
	m := &microserviceExecutor{}
	m.namespace = namespace
	m.name = name
	return m
}

func (ms *microserviceExecutor) Execute() error {
	microservice, err := config.GetMicroservice(ms.namespace, ms.name)
	if err != nil {
		return err
	}
	if err = print(microservice); err != nil {
		return err
	}
	return nil
}
