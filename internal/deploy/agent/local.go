package deployagent

import (
	"fmt"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"os/user"
)

type localExecutor struct {
	opt *Options
}

func newLocalExecutor(opt *Options) *localExecutor {
	exe := &localExecutor{}
	exe.opt = opt

	return exe
}

func (exe *localExecutor) Execute() error {
	// TODO (Serge) Execute back-end logic

	currUser, err := user.Current()
	if err != nil {
		return err
	}
	// Update configuration
	configEntry := config.Agent{
		Name: exe.opt.Name,
		User: currUser.Username,
		Host: "localhost",
	}
	err = config.AddAgent(exe.opt.Namespace, configEntry)
	if err != nil {
		return err
	}

	fmt.Printf("\nAgent %s/%s successfully deployed.\n", exe.opt.Namespace, exe.opt.Name)

	return config.Flush()
}
