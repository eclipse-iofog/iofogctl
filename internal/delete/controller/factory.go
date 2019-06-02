package deletecontroller

import (
	"github.com/eclipse-iofog/cli/internal/config"
)

type Executor interface {
	Execute() error
}

func NewExecutor(namespace, name string) (Executor, error) {
	// Get controller from config
	ctrl, err := config.GetController(namespace, name)
	if err != nil {
		return nil, err
	}

	// Local executor
	if ctrl.Host == "localhost" {
		return newLocalExecutor(namespace, name), nil
	}

	// Kubernetes executor
	if ctrl.KubeConfig != "" {
		return newKubernetesExecutor(namespace, name), nil
	}

	// Default executor
	return newRemoteExecutor(namespace, name), nil
}
