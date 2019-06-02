package get

import (
	"github.com/eclipse-iofog/cli/pkg/util"
)

type Executor interface {
	Execute() error
}

func NewExecutor(resourceType, namespace string) (Executor, error) {

	switch resourceType {
	case "namespaces":
		return newNamespaceExecutor(), nil
	case "controllers":
		return newControllerExecutor(namespace), nil
	case "agents":
		return newAgentExecutor(namespace), nil
	case "microservices":
		return newMicroserviceExecutor(namespace), nil
	default:
		msg := "Unknown resource: '" + resourceType + "'"
		return nil, util.NewInputError(msg)
	}
}
