package describe

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
)

type controllerExecutor struct {
	namespace string
	name      string
}

func newControllerExecutor(namespace, name string) *controllerExecutor {
	c := &controllerExecutor{}
	c.namespace = namespace
	c.name = name
	return c
}

func (exe *controllerExecutor) Execute() error {
	controller, err := config.GetController(exe.namespace, exe.name)
	if err != nil {
		return err
	}
	if err = print(controller); err != nil {
		return err
	}
	return nil
}
