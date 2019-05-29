package deploycontroller

import (
	"fmt"
	"github.com/eclipse-iofog/cli/internal/config"
)

type remoteExecutor struct {
	configManager *config.Manager
	opt           *Options
}

func newRemoteExecutor(opt *Options) *remoteExecutor {
	d := &remoteExecutor{}
	d.configManager = config.NewManager()
	d.opt = opt
	return d
}

func (exe *remoteExecutor) Execute(namespace, name string) error {
	// TODO (Serge) Execute back-end logic

	// Update configuration
	configEntry := config.Controller{
		Name:    name,
		User:    exe.opt.User,
		Host:    exe.opt.Host,
		KeyFile: exe.opt.KeyFile,
	}
	err := exe.configManager.AddController(namespace, configEntry)

	// TODO (Serge) Handle config file error, retry..?

	if err == nil {
		fmt.Printf("\nController %s/%s successfully deployed.\n", namespace, name)
	}
	return err
}
