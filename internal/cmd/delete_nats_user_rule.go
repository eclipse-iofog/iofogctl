package cmd

import (
	deletenatsuserrule "github.com/eclipse-iofog/iofogctl/internal/delete/natsuserrule"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteNatsUserRuleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "nats-user-rule NAME",
		Short:   "Delete a NATS user rule",
		Long:    `Delete a NATS user rule from the Controller.`,
		Example: `iofogctl delete nats-user-rule NAME`,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			exe, err := deletenatsuserrule.NewExecutor(namespace, name)
			util.Check(err)
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully deleted NATS user rule " + name)
		},
	}

	return cmd
}
