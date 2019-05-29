package deletecontroller

import (
	"github.com/eclipse-iofog/cli/internal/config"
)

type Executor interface {
	Execute() error
}

func NewExecutor(namespace, name string) (Executor, error) {
	// Instantiate config manager
	cfg := config.NewManager()

	// Find the requested controller
	ctrl, err := cfg.GetController(namespace, name)
	if err != nil {
		return nil, err
	}

	// Local executor
	if ctrl.Host == "localhost" {
		return newLocalExecutor(cfg, namespace, ctrl), nil
	}

	// Kubernetes executor
	if ctrl.KubeConfig != "" {
		return newKubernetesExecutor(cfg, namespace, ctrl), nil
	}

	// Default executor
	return newRemoteExecutor(cfg, namespace, ctrl), nil
}
