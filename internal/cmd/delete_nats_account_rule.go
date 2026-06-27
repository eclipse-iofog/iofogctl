package cmd

import (
	deletenatsaccountrule "github.com/eclipse-iofog/iofogctl/internal/delete/natsaccountrule"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteNatsAccountRuleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "nats-account-rule NAME",
		Short:   "Delete a NATS account rule",
		Long:    `Delete a NATS account rule from the Controller.`,
		Example: ex(`%[1]s delete nats-account-rule NAME`),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			exe, err := deletenatsaccountrule.NewExecutor(namespace, name)
			util.Check(err)
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully deleted NATS account rule " + name)
		},
	}

	return cmd
}
