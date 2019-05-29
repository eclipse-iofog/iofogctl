package deployagent

import (
	"fmt"
	"github.com/eclipse-iofog/cli/internal/config"
)

type remoteExecutor struct {
	configManager *config.Manager
	opt           *Options
}

func newRemoteExecutor(opt *Options) *remoteExecutor {
	exe := &remoteExecutor{}
	exe.configManager = config.NewManager()
	exe.opt = opt

	return exe
}

func (exe *remoteExecutor) Execute(namespace, name string) error {
	// TODO (Serge) Execute back-end logic

	// Update configuration
	configEntry := config.Agent{
		Name:    name,
		User:    exe.opt.User,
		Host:    exe.opt.Host,
		KeyFile: exe.opt.KeyFile,
	}
	err := exe.configManager.AddAgent(namespace, configEntry)

	// TODO (Serge) Handle config file error, retry..?

	if err == nil {
		fmt.Printf("\nAgent %s/%s successfully deployed.\n", namespace, name)
	}
	return err
}
