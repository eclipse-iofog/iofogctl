package describe

import (
	"github.com/eclipse-iofog/cli/pkg/util"
)

type Executor interface {
	Execute() error
}

func NewExecutor(resourceType, namespace, name string) (Executor, error) {
	switch resourceType {
	case "namespace":
		return newNamespaceExecutor(namespace), nil
	case "controller":
		return newControllerExecutor(namespace, name), nil
	case "agent":
		return newAgentExecutor(namespace, name), nil
	case "microservice":
		return newMicroserviceExecutor(namespace, name), nil
	default:
		msg := "Unknown resourceType: '" + resourceType + "'"
		return nil, util.NewInputError(msg)
	}
}
