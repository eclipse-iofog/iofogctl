package cmd

import (
	deleteserviceaccount "github.com/eclipse-iofog/iofogctl/internal/delete/serviceaccount"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteServiceAccountCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "serviceaccount APPLICATION_NAME/SERVICE_ACCOUNT_NAME",
		Short:   "Delete a ServiceAccount",
		Long:    `Delete a ServiceAccount from the Controller. ServiceAccounts are application-scoped; use APPLICATION_NAME/SERVICE_ACCOUNT_NAME (e.g. myapp/my-sa).`,
		Example: `iofogctl delete serviceaccount myapp/my-sa`,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			exe, err := deleteserviceaccount.NewExecutor(namespace, name)
			util.Check(err)
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully deleted serviceaccount " + name)
		},
	}

	return cmd
}
