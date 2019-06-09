package cmd

import (
	delete "github.com/eclipse-iofog/iofogctl/internal/delete/microservice"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteMicroserviceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "microservice NAME [AGENTNAME]",
		Short: "Delete a Microservice",
		Long:  `Delete a Microservice`,
		Example: `iofogctl delete microservice NAME
iofogctl delete microservice NAME`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get microservice name
			name := args[0]
			// Get namespace option
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Execute command
			microservice := delete.New()
			err = microservice.Execute(namespace, name)
			util.Check(err)
		},
	}

	return cmd
}
