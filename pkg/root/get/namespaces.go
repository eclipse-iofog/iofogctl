package get

import (
	"github.com/eclipse-iofog/cli/pkg/config"
)

type namespaceExecutor struct {
	configManager *config.Manager
}

func newNamespaceExecutor() *namespaceExecutor {
	n := &namespaceExecutor{}
	n.configManager = config.NewManager()
	return n
}

func (ns *namespaceExecutor) execute(string) error {
	namespaces := ns.configManager.GetNamespaces()
	rows := make([]row, len(namespaces))
	for idx, ns := range namespaces {
	    rows[idx].name = ns.Name
	    // TODO: (Serge) Get runtime info
	    rows[idx].status = "Active"
	    rows[idx].age = "-"
	}
	err := print(rows)
	return err
}