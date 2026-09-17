package cmd

import (
	deletemicroservicetemplate "github.com/eclipse-iofog/iofogctl/internal/delete/microservicetemplate"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteMicroserviceTemplateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "microservice-template NAME",
		Short:   "Delete a Microservice Template",
		Long:    `Delete a Microservice Template from the Controller.`,
		Example: ex(`%[1]s delete microservice-template NAME`),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			exe, err := deletemicroservicetemplate.NewExecutor(namespace, name)
			util.Check(err)
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully deleted microservice-template " + name)
		},
	}

	return cmd
}
