package cmd

import (
	delete "github.com/eclipse-iofog/iofogctl/internal/delete/volume"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteVolumeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "volume NAME",
		Short: "Delete an Volume",
		Long: `Delete an Volume.

The Volume will be deleted from the Agents that it is stored on.`,
		Example: ex(`%[1]s delete volume NAME`),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace of volume
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Run the command
			exe, err := delete.NewExecutor(namespace, name)
			util.Check(err)
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully deleted " + namespace + "/" + name)
		},
	}

	return cmd
}
