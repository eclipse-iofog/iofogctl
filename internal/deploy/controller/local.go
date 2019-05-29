package deploycontroller

import (
	"fmt"
	"github.com/eclipse-iofog/cli/internal/config"
	"os/user"
)

type localExecutor struct {
	configManager *config.Manager
	opt           *Options
}

func newLocalExecutor(opt *Options) *localExecutor {
	l := &localExecutor{}
	l.configManager = config.NewManager()
	l.opt = opt
	return l
}

func (exe *localExecutor) Execute(namespace, name string) error {
	// TODO (Serge) Execute back-end logic

	currUser, err := user.Current()
	if err != nil {
		return err
	}
	// Update configuration
	configEntry := config.Controller{
		Name: name,
		User: currUser.Username,
		Host: "localhost",
	}
	err = exe.configManager.AddController(namespace, configEntry)

	// TODO (Serge) Handle config file error, retry..?

	if err == nil {
		fmt.Printf("\nController %s/%s successfully deployed.\n", namespace, name)
	}
	return err
}
