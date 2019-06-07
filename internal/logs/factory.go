package logs

import (
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Executor interface {
	Execute() error
}

func NewExecutor(resourceType, namespace, name string) (Executor, error) {
	switch resourceType {
	case "controller":
		return newControllerExecutor(namespace, name), nil
	case "agent":
		return newAgentExecutor(namespace, name), nil
	case "microservice":
		return newMicroserviceExecutor(namespace, name), nil
	default:
		msg := "Unknown resource: '" + resourceType + "'"
		return nil, util.NewInputError(msg)
	}
}
