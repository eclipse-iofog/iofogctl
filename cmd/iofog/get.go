package main

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
		Example: `iofog get controllers
iofog get agents
iofog get microservices`,
		ValidArgs: []string{"namespaces", "controllers", "agents", "microservices"},
		Args:      cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Perform get for specified resource
			resource := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			exe, err := get.NewExecutor(resource)
			util.Check(err)

			err = exe.Execute(namespace)
			util.Check(err)
		},
	}

	return cmd
}
