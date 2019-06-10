package deploycontroller

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Executor interface {
	Execute() error
}

type Options struct {
	Name       string
	Namespace  string
	User       string
	Host       string
	KeyFile    string
	Local      bool
	KubeConfig string
	ImagesFile string
	Images     map[string]string
}

func NewExecutor(opt *Options) (Executor, error) {
	// Check the namespace exists
	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return nil, err
	}

	// Check controller already exists
	if len(ns.Controllers) > 0 {
		return nil, util.NewConflictError("Controller already exists in namespace " + opt.Namespace)
	}

	// Local executor
	if opt.Local == true {
		return newLocalExecutor(opt), nil
	}

	// Kubernetes executor
	if opt.KubeConfig != "" {
		// If image file specified, read it
		if opt.ImagesFile != "" {
			opt.Images = make(map[string]string)
			err = util.UnmarshalYAML(opt.ImagesFile, opt.Images)
			if err != nil {
				return nil, err
			}
		}
		return newKubernetesExecutor(opt), nil
	}

	// Default executor
	if opt.Host == "" || opt.KeyFile == "" || opt.User == "" {
		return nil, util.NewInputError("Must specify user, host, and key file flags for remote deployment")
	}
	return newRemoteExecutor(opt), nil
}
