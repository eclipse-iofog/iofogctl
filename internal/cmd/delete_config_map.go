package cmd

import (
	deleteconfigmap "github.com/eclipse-iofog/iofogctl/internal/delete/configmap"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteConfigMapCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "configmap NAME",
		Short:   "Delete a ConfigMap",
		Long:    `Delete a ConfigMap from the Controller.`,
		Example: ex(`%[1]s delete configmap NAME`),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Get an executor for the command
			exe, err := deleteconfigmap.NewExecutor(namespace, name)
			util.Check(err)
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully deleted configmap " + name)
		},
	}

	return cmd
}
