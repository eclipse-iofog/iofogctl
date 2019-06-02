package describe

import (
	"github.com/eclipse-iofog/cli/internal/config"
)

type namespaceExecutor struct {
	name string
}

func newNamespaceExecutor(name string) *namespaceExecutor {
	n := &namespaceExecutor{}
	n.name = name
	return n
}

func (exe *namespaceExecutor) Execute() error {
	namespace, err := config.GetNamespace(exe.name)
	if err != nil {
		return err
	}
	if err = print(namespace); err != nil {
		return err
	}
	return nil
}
