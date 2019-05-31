package main

import (
	delete "github.com/eclipse-iofog/cli/internal/delete/controller"
	"github.com/eclipse-iofog/cli/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteControllerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "controller name",
		Short:   "Delete a Controller",
		Long:    `Delete a Controller`,
		Example: `iofogctl delete controller my_controller_name`,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace of controller
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Get an executor for the command
			exe, err := delete.NewExecutor(namespace, name)
			util.Check(err)

			// Run the command
			err = exe.Execute()
			util.Check(err)
		},
	}

	return cmd
}
