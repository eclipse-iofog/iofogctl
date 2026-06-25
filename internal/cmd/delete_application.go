package cmd

import (
	delete "github.com/eclipse-iofog/iofogctl/internal/delete/application"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteApplicationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "application NAME",
		Short:   "Delete an application",
		Long:    `Delete an application and all its components`,
		Example: ex(`%[1]s delete application NAME`),
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get microservice name
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Execute command
			err = delete.Execute(namespace, name)
			util.Check(err)

			util.PrintSuccess("Successfully deleted " + namespace + "/" + name)
		},
	}

	return cmd
}
