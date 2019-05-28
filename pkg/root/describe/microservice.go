package describe

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
	microservice, err := ns.configManager.GetMicroservice(namespace, name)
	if err != nil {
		return err
	}
	if err = print(microservice); err != nil {
		return err
	}
	return nil
}
