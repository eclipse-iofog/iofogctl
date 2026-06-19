package configure

import (
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Options struct {
	ResourceType string
	Namespace    string
	Name         string
	KubeConfig   string
	KeyFile      string
	User         string
	Port         int
	UseDetached  bool
}

var multipleResources = map[string]bool{
	"agents":      true,
	"controllers": true,
}

func NewExecutor(opt *Options) (execute.Executor, error) {
	switch opt.ResourceType {
	case "current-namespace":
		return newDefaultNamespaceExecutor(opt), nil
	case "default-namespace":
		return newDefaultNamespaceExecutor(opt), nil
	case "controlplane":
		return newControlPlaneExecutor(opt), nil
	case "controller":
		return newControllerExecutor(opt), nil
	case "agent":
		return newAgentExecutor(opt), nil
	default:
		if _, exists := multipleResources[opt.ResourceType]; !exists {
			return nil, util.NewInputError("Unsupported resource: " + opt.ResourceType)
		}
		return newMultipleExecutor(opt), nil
	}
}
