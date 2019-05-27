package get

import (
	"github.com/eclipse-iofog/cli/pkg/config"
)

type microserviceExecutor struct {
	configManager *config.Manager
}

func newMicroserviceExecutor() *microserviceExecutor {
	a := &microserviceExecutor{}
	a.configManager = config.NewManager()
	return a
}

func (ms *microserviceExecutor) execute(namespace string) error {
	microservices, err := ms.configManager.GetMicroservices(namespace)
	if err != nil {
		return err
	}
	rows := make([]row, len(microservices))
	for idx, ms := range microservices {
		rows[idx].name = ms.Name
		// TODO: (Serge) Get runtime info
		rows[idx].status = "-"
		rows[idx].age = "-"
	}
	err = print(rows)
	return err
}