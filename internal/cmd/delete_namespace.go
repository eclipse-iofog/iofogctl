package cmd

import (
	delete "github.com/eclipse-iofog/iofogctl/internal/delete/namespace"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteNamespaceCommand() *cobra.Command {
	force := false
	cmd := &cobra.Command{
		Use:   "namespace NAME",
		Short: "Delete a Namespace",
		Long: `Delete a Namespace.

The Namespace must be empty.

If you would like to delete all resources in the Namespace, use the --force flag.`,
		Example: `iofogctl delete namespace NAME`,
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get microservice name
			name := args[0]

			// Execute command
			err := delete.Execute(name, force)
			util.Check(err)

			util.PrintSuccess("Successfully deleted Namespace " + name)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force deletion of all resources within the Namespace")

	return cmd
}
