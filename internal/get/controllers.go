package get

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

func (ctrl *controllerExecutor) Execute(namespace string) error {
	controllers, err := ctrl.configManager.GetControllers(namespace)
	if err != nil {
		return err
	}

	rows := make([]row, len(controllers))
	for idx, ctrl := range controllers {
		rows[idx].name = ctrl.Name
		// TODO: (Serge) Get runtime info
		rows[idx].status = "-"
		rows[idx].age = "-"
	}
	err = print(rows)
	return err
}
