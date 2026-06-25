package cmd

import (
	deleteservice "github.com/eclipse-iofog/iofogctl/internal/delete/service"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteServiceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "service NAME",
		Short:   "Delete a Service",
		Long:    `Delete a Service from the Controller.`,
		Example: ex(`%[1]s delete service NAME`),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Get an executor for the command
			exe, err := deleteservice.NewExecutor(namespace, name)
			util.Check(err)
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully deleted secret " + name)
		},
	}

	return cmd
}
