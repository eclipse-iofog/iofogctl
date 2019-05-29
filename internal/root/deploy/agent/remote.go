package deployagent

import (
	"fmt"
	"github.com/eclipse-iofog/cli/internal/config"
)

type remoteExecutor struct {
	configManager *config.Manager
	opt           *options
}

func newRemoteExecutor(opt *options) *remoteExecutor {
	exe := &remoteExecutor{}
	exe.configManager = config.NewManager()
	exe.opt = opt

	return exe
}

func (exe *remoteExecutor) execute(namespace, name string) error {
	// TODO (Serge) Execute back-end logic

	// Update configuration
	configEntry := config.Agent{
		Name:    name,
		User:    exe.opt.user,
		Host:    exe.opt.host,
		KeyFile: exe.opt.keyFile,
	}
	err := exe.configManager.AddAgent(namespace, configEntry)

	// TODO (Serge) Handle config file error, retry..?

	if err == nil {
		fmt.Printf("\nAgent %s/%s successfully deployed.\n", namespace, name)
	}
	return err
}
