package get

import (
	"github.com/eclipse-iofog/cli/pkg/util"
)

type Executor interface {
	Execute(string) error
}

func NewExecutor(resource string) (Executor, error) {

	switch resource {
	case "namespaces":
		return newNamespaceExecutor(), nil
	case "controllers":
		return newControllerExecutor(), nil
	case "agents":
		return newAgentExecutor(), nil
	case "microservices":
		return newMicroserviceExecutor(), nil
	default:
		msg := "Unknown resource: '" + resource + "'"
		return nil, util.NewInputError(msg)
	}
}
