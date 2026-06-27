package cmd

import (
	delete "github.com/eclipse-iofog/iofogctl/internal/delete/registry"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteRegistryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "registry ID",
		Short:   "Delete a Registry",
		Long:    `Delete a Registry from the Controller.`,
		Example: ex(`%[1]s delete registry ID`),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace
			id := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Get an executor for the command
			exe, err := delete.NewExecutor(namespace, id)
			util.Check(err)
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully deleted registry " + id)
		},
	}

	return cmd
}
