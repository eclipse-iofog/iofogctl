package cmd

import (
	delete "github.com/eclipse-iofog/iofogctl/internal/delete/namespace"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteNamespaceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "namespace name",
		Short:   "Delete a Namespace",
		Long:    `Delete a Namespace`,
		Example: `iofogctl delete namespace my_namespace_name`,
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get microservice name
			name := args[0]

			// Execute command
			err := delete.Execute(name)
			util.Check(err)
		},
	}

	return cmd
}
