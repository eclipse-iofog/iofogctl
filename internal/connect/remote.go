package connect

type remoteExecutor struct {
	opt *Options
}

func newRemoteExecutor(opt *Options) *remoteExecutor {
	r := &remoteExecutor{}
	r.opt = opt
	return r
}

func (exe *remoteExecutor) Execute() (err error) {
	return nil
}
