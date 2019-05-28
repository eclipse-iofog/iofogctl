package deletecontroller

import (
	"fmt"
	"github.com/eclipse-iofog/cli/pkg/config"
)

type localExecutor struct {
	configManager *config.Manager
	namespace     string
	controller    config.Controller
}

func newLocalExecutor(cfg *config.Manager, ns string, ctrl config.Controller) *localExecutor {
	exe := &localExecutor{}
	exe.configManager = cfg
	exe.namespace = ns
	exe.controller = ctrl
	return exe
}

func (exe *localExecutor) execute() error {
	// TODO (Serge) Execute back-end logic

	// Update configuration
	err := exe.configManager.DeleteController(exe.namespace, exe.controller.Name)

	// TODO (Serge) Handle config file error, retry..?

	if err == nil {
		fmt.Printf("\nController %s/%s successfully deleted.\n", exe.namespace, exe.controller.Name)
	}
	return err
}
