package deleteagent

import (
	"github.com/eclipse-iofog/cli/pkg/config"
)

type executor interface {
	execute() error
}

func getExecutor(namespace, name string) (executor, error) {
	// Instantiate config manager
	cfg := config.NewManager()

	// Find the requested controller
	ctrl, err := cfg.GetAgent(namespace, name)
	if err != nil {
		return nil, err
	}

	// Local executor
	if ctrl.Host == "localhost" {
		return newLocalExecutor(cfg, namespace, ctrl), nil
	}

	// Default executor
	return newRemoteExecutor(cfg, namespace, ctrl), nil
}
