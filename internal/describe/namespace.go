package describe

import (
	"github.com/eclipse-iofog/cli/internal/config"
)

type namespaceExecutor struct {
	configManager *config.Manager
}

func newNamespaceExecutor() *namespaceExecutor {
	n := &namespaceExecutor{}
	n.configManager = config.NewManager()
	return n
}

func (ns *namespaceExecutor) Execute(name string, empty string) error {
	namespace, err := ns.configManager.GetNamespace(name)
	if err != nil {
		return err
	}
	if err = print(namespace); err != nil {
		return err
	}
	return nil
}
