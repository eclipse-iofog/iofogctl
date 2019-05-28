package deploycontroller

import (
	"github.com/eclipse-iofog/cli/pkg/util"
)

type executor interface {
	execute(string, string) error
}

func getExecutor(opt *options) (executor, error) {
	// Local executor
	if *opt.local == true {
		return newLocalExecutor(opt), nil
	}

	// Kubernetes executor
	if *opt.kubeConfig != "" {
		return newKubernetesExecutor(opt), nil
	}

	// Default executor
	if *opt.host == "" || *opt.keyFile == "" || *opt.user == "" {
		return nil, util.NewInputError("Must specify user, host, and key file flags for remote deployment")
	}
	return newDefaultExecutor(opt), nil
}
