package cmd

import (
	deleterole "github.com/eclipse-iofog/iofogctl/internal/delete/role"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteRoleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "role NAME",
		Short:   "Delete a Role",
		Long:    `Delete a Role from the Controller.`,
		Example: `iofogctl delete role NAME`,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			exe, err := deleterole.NewExecutor(namespace, name)
			util.Check(err)
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully deleted role " + name)
		},
	}

	return cmd
}
