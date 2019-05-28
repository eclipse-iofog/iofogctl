package deployagent

import (
	"fmt"
	"github.com/eclipse-iofog/cli/pkg/config"
	"os/user"
)

type localExecutor struct {
	configManager *config.Manager
	opt           *options
}

func newLocalExecutor(opt *options) *localExecutor {
	exe := &localExecutor{}
	exe.configManager = config.NewManager()
	exe.opt = opt

	return exe
}

func (exe *localExecutor) execute(namespace, name string) error {
	// TODO (Serge) Execute back-end logic

	currUser, err := user.Current()
	if err != nil {
		return err
	}
	// Update configuration
	configEntry := config.Agent{
		Name: name,
		User: currUser.Username,
		Host: "localhost",
	}
	err = exe.configManager.AddAgent(namespace, configEntry)

	// TODO (Serge) Handle config file error, retry..?

	if err == nil {
		fmt.Printf("\nAgent %s/%s successfully deployed.\n", namespace, name)
	}
	return err
}
