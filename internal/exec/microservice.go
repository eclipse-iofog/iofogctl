package exec

import (
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
)

type microserviceExecutor struct {
	namespace string
	name      string
}

func newMicroserviceExecutor(namespace, name string) *microserviceExecutor {
	return &microserviceExecutor{
		namespace: namespace,
		name:      name,
	}
}

func (exe *microserviceExecutor) GetName() string {
	return exe.name
}

func (exe *microserviceExecutor) Execute() error {
	label := "Microservice " + exe.name
	return runExecSession(exe.namespace, label, func(clt *client.Client) (*client.ExecSession, error) {
		return dialMicroserviceExec(clt, exe.name)
	})
}
