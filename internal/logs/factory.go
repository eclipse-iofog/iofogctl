package logs

import (
	"github.com/eclipse-iofog/cli/pkg/util"
)

type Executor interface {
	Execute(string, string) error
}

func NewExecutor(resource string) (Executor, error) {
	switch resource {
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
