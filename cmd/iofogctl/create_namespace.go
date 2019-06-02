package main

import (
	create "github.com/eclipse-iofog/cli/internal/create/namespace"
	"github.com/eclipse-iofog/cli/pkg/util"
	"github.com/spf13/cobra"
)

func newCreateNamespaceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "namespace name",
		Short:   "Create a Namespace",
		Long:    `Create a Namespace`,
		Example: `iofogctl create namespace my_namespace_name`,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace of agent
			name := args[0]

			// Run the command
			err := create.Execute(name)
			util.Check(err)
		},
	}

	return cmd
}
