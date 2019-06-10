package deploycontroller

import (
	"fmt"
	"github.com/eclipse-iofog/iofogctl/internal/config"
)

type remoteExecutor struct {
	opt *Options
}

func newRemoteExecutor(opt *Options) *remoteExecutor {
	d := &remoteExecutor{}
	d.opt = opt
	return d
}

func (exe *remoteExecutor) Execute() error {
	// TODO (Serge) Execute back-end logic

	// Update configuration
	configEntry := config.Controller{
		Name:    exe.opt.Name,
		User:    exe.opt.User,
		Host:    exe.opt.Host,
		KeyFile: exe.opt.KeyFile,
	}
	err := config.AddController(exe.opt.Namespace, configEntry)
	if err != nil {
		return err
	}

	fmt.Printf("\nController %s/%s successfully deployed.\n", exe.opt.Namespace, exe.opt.Name)

	return config.Flush()
}
