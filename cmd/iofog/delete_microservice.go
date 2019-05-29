package main

import (
	delete "github.com/eclipse-iofog/cli/internal/delete/microservice"
	"github.com/eclipse-iofog/cli/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteMicroserviceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "microservice name [agent]",
		Short: "Delete a Microservice",
		Long:  `Delete a Microservice`,
		Example: `iofog delete microservice my_microservice_name
iofog delete microservice my_microservice_name`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Execute
			microservice := delete.New()
			err = microservice.Execute(namespace, name)
			util.Check(err)
		},
	}

	return cmd
}
