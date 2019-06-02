package get

import (
	"github.com/eclipse-iofog/cli/internal/config"
)

type namespaceExecutor struct {
}

func newNamespaceExecutor() *namespaceExecutor {
	n := &namespaceExecutor{}
	return n
}

func (exe *namespaceExecutor) Execute() error {
	namespaces := config.GetNamespaces()
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
