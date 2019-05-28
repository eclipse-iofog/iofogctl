package deploycontroller

import (
	"fmt"
	"github.com/eclipse-iofog/cli/pkg/config"
)

type defaultExecutor struct {
	configManager *config.Manager
	opt           *options
}

func newDefaultExecutor(opt *options) *defaultExecutor {
	d := &defaultExecutor{}
	d.configManager = config.NewManager()
	d.opt = opt
	return d
}

func (exe *defaultExecutor) execute(namespace, name string) error {
	// TODO (Serge) Execute back-end logic

	// Update configuration
	configEntry := config.Controller{
		Name:    name,
		User:    *exe.opt.user,
		Host:    *exe.opt.host,
		KeyFile: *exe.opt.keyFile,
	}
	err := exe.configManager.AddController(namespace, configEntry)

	// TODO (Serge) Handle config file error, retry..?

	if err == nil {
		fmt.Printf("\nController %s/%s successfully deployed.\n", namespace, name)
	}
	return err
}
