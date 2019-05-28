package deploycontroller

import (
	"fmt"
	"github.com/eclipse-iofog/cli/internal/config"
)

type remoteExecutor struct {
	configManager *config.Manager
	opt           *options
}

func newRemoteExecutor(opt *options) *remoteExecutor {
	d := &remoteExecutor{}
	d.configManager = config.NewManager()
	d.opt = opt
	return d
}

func (exe *remoteExecutor) execute(namespace, name string) error {
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
