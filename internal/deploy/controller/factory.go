package deploycontroller

import (
	"github.com/eclipse-iofog/cli/pkg/util"
)

type Executor interface {
	Execute(string, string) error
}

type Options struct {
	User       string
	Host       string
	KeyFile    string
	Local      bool
	KubeConfig string
}

func NewExecutor(opt *Options) (Executor, error) {
	// Local executor
	if opt.Local == true {
		return newLocalExecutor(opt), nil
	}

	// Kubernetes executor
	if opt.KubeConfig != "" {
		return newKubernetesExecutor(opt), nil
	}

	// Default executor
	if opt.Host == "" || opt.KeyFile == "" || opt.User == "" {
		return nil, util.NewInputError("Must specify user, host, and key file flags for remote deployment")
	}
	return newRemoteExecutor(opt), nil
}
