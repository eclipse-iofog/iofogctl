package rollback

import (
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Options struct {
	ResourceType string
	Namespace    string
	Name         string
	Semver       string
}

func NewExecutor(opt Options) (execute.Executor, error) {
	switch opt.ResourceType {
	case "agent":
		return newAgentExecutor(opt), nil
	default:
		return nil, util.NewInputError("Unsupported resource: " + opt.ResourceType)
	}
}
