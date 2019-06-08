package connect

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Options struct {
	Namespace string
	Name      string
	Host      string
	KubeFile  string
}

type Executor interface {
	Execute() error
}

func NewExecutor(opt *Options) (Executor, error) {
	// Check namespace is empty
	ns, err := config.GetNamespace(opt.Namespace)
	if err == nil {
		if len(ns.Agents) != 0 || len(ns.Controllers) != 0 || len(ns.Microservices) != 0 {
			return nil, util.NewConflictError("You must use an empty namespace to connect to existing ioFog services")
		}
	}

	if opt.KubeFile != "" {
		return newKubernetesExecutor(opt), nil
	}

	if opt.Host == "" {
		return nil, util.NewInputError("Must specify Controller IP if connecting to non-Kubernetes Controller")
	}

	return newRemoteExecutor(opt), nil
}
