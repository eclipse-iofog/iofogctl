package configure

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type defaultNamespaceExecutor struct {
	name string
}

func newDefaultNamespaceExecutor(opt *Options) *defaultNamespaceExecutor {
	return &defaultNamespaceExecutor{
		name: opt.Name,
	}
}

func (exe *defaultNamespaceExecutor) GetName() string {
	return exe.name
}

func (exe *defaultNamespaceExecutor) Execute() error {
	if exe.name == "" {
		return util.NewInputError("Must specify Namespace")
	}
	return config.SetDefaultNamespace(exe.name)
}
