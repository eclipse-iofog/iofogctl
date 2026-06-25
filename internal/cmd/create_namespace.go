package cmd

import (
	create "github.com/eclipse-iofog/iofogctl/internal/create/namespace"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newCreateNamespaceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "namespace NAME",
		Short: "Create a Namespace",
		Long: ex(`Create a Namespace.

A Namespace contains all components of an Edge Compute Network.

A single instance of %[1]s can be used to manage any number of Edge Compute Networks.`),
		Example: ex(`%[1]s create namespace NAME`),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace of agent
			name := args[0]

			// Run the command
			err := create.Execute(name)
			util.Check(err)

			util.PrintSuccess("Successfully created namespace " + name)
		},
	}

	return cmd
}
