package cmd

import (
	deleteruntimeclass "github.com/eclipse-iofog/iofogctl/internal/delete/runtimeclass"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteRuntimeClassCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "runtimeclass NAME",
		Short:   "Delete a RuntimeClass",
		Long:    `Delete a RuntimeClass from the Controller.`,
		Example: ex(`%[1]s delete runtimeclass NAME`),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			exe, err := deleteruntimeclass.NewExecutor(namespace, name)
			util.Check(err)
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully deleted runtimeclass " + name)
		},
	}

	return cmd
}
