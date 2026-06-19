package cmd

import (
	delete "github.com/eclipse-iofog/iofogctl/internal/delete/template"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteApplicationTemplateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "application-template NAME",
		Short:   "Delete an application-template",
		Long:    `Delete an application-template`,
		Example: `iofogctl delete application-template NAME`,
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
