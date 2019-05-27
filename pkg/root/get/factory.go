package get

import (
	"github.com/eclipse-iofog/cli/pkg/util"
)

type executor interface {
	execute(string) error
}

func getExecutor(resource string) (executor, error) {

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