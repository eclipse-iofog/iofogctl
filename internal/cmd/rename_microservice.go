package cmd

import (
	rename "github.com/eclipse-iofog/iofogctl/internal/rename/microservice"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newRenameMicroserviceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "microservice NAME NEW_NAME",
		Short:   "Rename a Microservice",
		Long:    `Rename a Microservice`,
		Example: `iofogctl rename microservice NAME NEW_NAME`,
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and new name of the microservice
			name := args[0]
			newName := args[1]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Get an executor for the command
			err = rename.Execute(namespace, name, newName)
			util.Check(err)

			util.PrintSuccess(getRenameSuccessMessage("Microservice", name, newName))
		},
	}

	return cmd
}
