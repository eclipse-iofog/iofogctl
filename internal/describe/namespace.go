package describe

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type namespaceExecutor struct {
	name     string
	filename string
}

func newNamespaceExecutor(name, filename string) *namespaceExecutor {
	n := &namespaceExecutor{}
	n.name = name
	n.filename = filename
	return n
}

func (exe *namespaceExecutor) GetName() string {
	return exe.name
}

func (exe *namespaceExecutor) Execute() error {
	namespace, err := config.GetNamespace(exe.name)
	if err != nil {
		return err
	}
	if exe.filename == "" {
		if err := util.Print(namespace); err != nil {
			return err
		}
	} else {
		if err := util.FPrint(namespace, exe.filename); err != nil {
			return err
		}
	}
	return nil
}
