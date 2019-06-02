package get

import (
	"github.com/eclipse-iofog/cli/internal/config"
)

type controllerExecutor struct {
	namespace string
}

func newControllerExecutor(namespace string) *controllerExecutor {
	c := &controllerExecutor{}
	c.namespace = namespace
	return c
}

func (exe *controllerExecutor) Execute() error {
	controllers, err := config.GetControllers(exe.namespace)
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
