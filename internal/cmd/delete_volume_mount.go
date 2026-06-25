package cmd

import (
	deletevolumemount "github.com/eclipse-iofog/iofogctl/internal/delete/volumemount"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteVolumeMountCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "volume-mount NAME",
		Short:   "Delete a Volume Mount",
		Long:    `Delete a Volume Mount from the Controller.`,
		Example: ex(`%[1]s delete volume-mount NAME`),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Get an executor for the command
			exe, err := deletevolumemount.NewExecutor(namespace, name)
			util.Check(err)
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully deleted secret " + name)
		},
	}

	return cmd
}
