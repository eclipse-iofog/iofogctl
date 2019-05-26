package get

import (
	"fmt"
	"github.com/eclipse-iofog/cli/pkg/config"
	"github.com/eclipse-iofog/cli/pkg/util"
	"text/tabwriter"
	"os"
)
type get struct {
	configManager *config.Manager
}

func new() *get {
	g := &get{}
	g.configManager = config.NewManager()
	return g
}

type row struct {
	name string
	status string
	age string
}

func (get *get) execute(resource, namespace string) error {

	switch resource {

	case "namespaces":
		namespaces := get.configManager.GetNamespaces()
		rows := make([]row, len(namespaces))
		for idx, ns := range namespaces {
			rows[idx].name = ns.Name
			// TODO: (Serge) Get runtime info
			rows[idx].status = "Active"
			rows[idx].age = "-"
		}
		print(rows)

	case "controllers":
		controllers, err := get.configManager.GetControllers(namespace)
		if err != nil {
			return err
		}

		rows := make([]row, len(controllers))
		for idx, ctrl := range controllers {
			rows[idx].name = ctrl.Name
			// TODO: (Serge) Get runtime info
			rows[idx].status = "-"
			rows[idx].age = "-"
		}
		print(rows)

	case "agents":
		agents, err := get.configManager.GetAgents(namespace)
		if err != nil {
			return err
		}
		rows := make([]row, len(agents))
		for idx, agent := range agents {
			rows[idx].name = agent.Name
			// TODO: (Serge) Get runtime info
			rows[idx].status = "-"
			rows[idx].age = "-"
		}
		print(rows)

	case "microservices":
		microservices, err := get.configManager.GetMicroservices(namespace)
		if err != nil {
			return err
		}
		rows := make([]row, len(microservices))
		for idx, ms := range microservices {
			rows[idx].name = ms.Name
			// TODO: (Serge) Get runtime info
			rows[idx].status = "-"
			rows[idx].age = "-"
		}
		print(rows)

	default:
		msg := "Unknown resource: '" + resource + "'"
		return util.NewInputError(msg)
	}

	return nil
}

func print(rows []row) {
	minWidth := 16
	tabWidth := 8
	padding := 0
	writer := tabwriter.NewWriter(os.Stdout, minWidth, tabWidth, padding, '\t', 0)
	defer writer.Flush()

	headers := [3]string{"NAME", "STATUS", "AGE"}
	fmt.Fprintf(writer, "\n%s\t%s\t%s\t\n", headers[0], headers[1], headers[2])

	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t\n", row.name, row.status, row.age)
	}
	fmt.Fprintf(writer, "\n")
}