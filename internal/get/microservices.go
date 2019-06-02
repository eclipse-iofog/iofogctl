package get

import (
	"github.com/eclipse-iofog/cli/internal/config"
)

type microserviceExecutor struct {
	namespace string
}

func newMicroserviceExecutor(namespace string) *microserviceExecutor {
	a := &microserviceExecutor{}
	a.namespace = namespace
	return a
}

func (exe *microserviceExecutor) Execute() error {
	microservices, err := config.GetMicroservices(exe.namespace)
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
