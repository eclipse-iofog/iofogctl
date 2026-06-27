package cmd

import (
	deletecertificate "github.com/eclipse-iofog/iofogctl/internal/delete/certificate"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteCertificateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "certificate NAME",
		Short:   "Delete a Certificate",
		Long:    `Delete a Certificate from the Controller.`,
		Example: ex(`%[1]s delete certificate NAME`),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Get an executor for the command
			exe, err := deletecertificate.NewExecutor(namespace, name)
			util.Check(err)
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully deleted certificate " + name)
		},
	}

	return cmd
}
