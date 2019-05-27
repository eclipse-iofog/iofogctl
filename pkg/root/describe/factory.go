package describe

import (
	"github.com/eclipse-iofog/cli/pkg/util"
)

type executor interface {
	execute(string, string) error
}

func getExecutor(resource string) (executor, error) {
	switch resource {
	case "namespace":
		return newNamespaceExecutor(), nil
	case "controller":
		return newControllerExecutor(), nil
	case "agent":
		return newAgentExecutor(), nil
	case "microservice":
		return newMicroserviceExecutor(), nil
	default:
		msg := "Unknown resource: '" + resource + "'"
		return nil, util.NewInputError(msg)
	}
}