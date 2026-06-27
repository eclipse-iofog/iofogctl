package get

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type namespaceExecutor struct {
}

func newNamespaceExecutor() *namespaceExecutor {
	n := &namespaceExecutor{}
	return n
}

func (exe *namespaceExecutor) GetName() string {
	return ""
}

func (exe *namespaceExecutor) Execute() error {
	namespacesNames := config.GetNamespaces()
	namespaces := make([]*rsc.Namespace, len(namespacesNames))
	for idx, n := range namespacesNames {
		ns, err := config.GetNamespace(n)
		if err != nil {
			return err
		}
		namespaces[idx] = ns
	}

	// Generate table and headers
	table := make([][]string, len(namespaces))
	headers := []string{"NAMESPACE", "AGE"}
	table[0] = append(table[0], headers...)

	// Populate rows
	for idx, ns := range namespaces {
		age, err := util.ElapsedUTC(ns.Created, util.NowUTC())
		if err != nil {
			age = "-"
		}
		row := []string{
			ns.Name,
			age,
		}
		if ns.Name == config.GetDefaultNamespaceName() {
			row[0] = ns.Name + "*"
			prepend := [][]string{table[0]}
			table = append(prepend, table...)
			table[1] = row
		} else {
			table[idx+1] = append(table[idx+1], row...)
		}
	}

	// Print the table
	return print(table)
}
