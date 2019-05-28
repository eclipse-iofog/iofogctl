package describe

import (
	"github.com/eclipse-iofog/cli/internal/config"
)

type controllerExecutor struct {
	configManager *config.Manager
}

func newControllerExecutor() *controllerExecutor {
	c := &controllerExecutor{}
	c.configManager = config.NewManager()
	return c
}

func (ns *controllerExecutor) execute(namespace string, name string) error {
	controller, err := ns.configManager.GetController(namespace, name)
	if err != nil {
		return err
	}
	if err = print(controller); err != nil {
		return err
	}
	return nil
}
