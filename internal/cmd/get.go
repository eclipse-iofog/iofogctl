package cmd

import (
	"github.com/eclipse-iofog/cli/internal/get"
	"github.com/eclipse-iofog/cli/pkg/util"
	"github.com/spf13/cobra"
)

func newGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get resource",
		Short: "Get information of existing resources",
		Long:  `Get information of existing resources`,
		Example: `iofogctl get all
iofogctl get namespaces
iofogctl get controllers
iofogctl get agents
iofogctl get microservices`,
		ValidArgs: []string{"namespaces", "all", "controllers", "agents", "microservices"},
		Args:      cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get resource type arg
			resource := args[0]

			// Get namespace option
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Get executor for get command
			exe, err := get.NewExecutor(resource, namespace)
			util.Check(err)

			// Execute the get command
			err = exe.Execute()
			util.Check(err)
		},
	}

	return cmd
}
