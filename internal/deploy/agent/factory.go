package deployagent

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Executor interface {
	Execute() error
}

type Options struct {
	Namespace string
	Name      string
	User      string
	Host      string
	KeyFile   string
	Local     bool
}

func NewExecutor(opt *Options) (Executor, error) {
	// Check the namespace exists
	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return nil, err
	}

	// Check Controller exists
	if len(ns.Controllers) != 1 {
		return nil, util.NewInputError("You must deploy a Controller before deploying Agents in this namespace")
	}

	// Check Agent already exists
	_, err = config.GetAgent(opt.Namespace, opt.Name)
	if err == nil {
		return nil, util.NewConflictError(opt.Namespace + "/" + opt.Name)
	}

	// Local executor
	if opt.Local == true {
		return newLocalExecutor(opt), nil
	}

	// Default executor
	if opt.Host == "" || opt.KeyFile == "" || opt.User == "" {
		return nil, util.NewInputError("Must specify user, host, and key file flags for remote deployment")
	}
	return newRemoteExecutor(opt), nil
}
