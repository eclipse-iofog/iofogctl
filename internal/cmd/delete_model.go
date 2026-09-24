package cmd

import (
	deletemodel "github.com/eclipse-iofog/iofogctl/internal/delete/model"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteModelCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "model NAME",
		Short:   "Delete a Model",
		Long:    `Delete a Model from the Controller.`,
		Example: ex(`%[1]s delete model NAME`),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			exe, err := deletemodel.NewExecutor(namespace, name)
			util.Check(err)
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully deleted model " + name)
		},
	}

	return cmd
}
