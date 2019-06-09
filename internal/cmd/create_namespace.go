package cmd

import (
	create "github.com/eclipse-iofog/iofogctl/internal/create/namespace"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newCreateNamespaceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "namespace NAME",
		Short:   "Create a Namespace",
		Long:    `Create a Namespace`,
		Example: `iofogctl create namespace NAME`,
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
