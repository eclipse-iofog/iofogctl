package connect

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
)

type remoteExecutor struct {
	opt *Options
}

func newRemoteExecutor(opt *Options) *remoteExecutor {
	r := &remoteExecutor{}
	r.opt = opt
	return r
}

func (exe *remoteExecutor) Execute() (err error) {
	// Establish connection
	err = connect(exe.opt, exe.opt.Endpoint)
	if err != nil {
		return err
	}
	return config.Flush()
}
