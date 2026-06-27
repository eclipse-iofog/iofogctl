package exec

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Options struct {
	Resource   string
	Name       string
	Namespace  string
	DebugImage *string
}

func NewExecutor(opt *Options) (execute.Executor, error) {
	switch opt.Resource {
	case "microservice":
		return newMicroserviceExecutor(opt.Namespace, opt.Name), nil
	case "agent":
		return newAgentExecutor(opt.Namespace, opt.Name, opt.DebugImage), nil
	default:
		return nil, util.NewInputError(fmt.Sprintf("Unknown resources: %s", opt.Resource))
	}
}
