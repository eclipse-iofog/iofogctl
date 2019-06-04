package deployagent

import (
	"fmt"
	"github.com/eclipse-iofog/cli/internal/config"
	"github.com/eclipse-iofog/cli/pkg/iofog"
)

type remoteExecutor struct {
	opt *Options
}

func newRemoteExecutor(opt *Options) *remoteExecutor {
	exe := &remoteExecutor{}
	exe.opt = opt

	return exe
}

func (exe *remoteExecutor) Execute() error {
	// Install the agent stack on the server
	agent := iofog.NewAgent(exe.opt.User, exe.opt.Host, exe.opt.KeyFile)
	err := agent.Bootstrap()
	if err != nil {
		return err
	}
	err = agent.Configure()
	if err != nil {
		return err
	}

	// Update configuration
	configEntry := config.Agent{
		Name:    exe.opt.Name,
		User:    exe.opt.User,
		Host:    exe.opt.Host,
		KeyFile: exe.opt.KeyFile,
	}
	err = config.AddAgent(exe.opt.Namespace, configEntry)
	if err != nil {
		return err
	}

	// TODO (Serge) Handle config file error, retry..?

	fmt.Printf("\nAgent %s/%s successfully deployed.\n", exe.opt.Namespace, exe.opt.Name)

	return nil
}
