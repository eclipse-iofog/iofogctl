package cmd

import (
	deletesecret "github.com/eclipse-iofog/iofogctl/internal/delete/secret"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteSecretCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "secret NAME",
		Short:   "Delete a Secret",
		Long:    `Delete a Secret from the Controller.`,
		Example: `iofogctl delete secret NAME`,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Get an executor for the command
			exe, err := deletesecret.NewExecutor(namespace, name)
			util.Check(err)
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully deleted secret " + name)
		},
	}

	return cmd
}
