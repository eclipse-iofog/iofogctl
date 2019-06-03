package cmd

import (
	deploy "github.com/eclipse-iofog/cli/internal/deploy/microservice"
	"github.com/eclipse-iofog/cli/pkg/util"
	"github.com/spf13/cobra"
)

func newDeployMicroserviceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "microservice name [agent]",
		Short: "Deploy a Microservice",
		Long:  `Deploy a Microservice`,
		Example: `iofogctl deploy microservice my_microservice_name
iofogctl deploy microservice my_microservice_name my_agent_name`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get microservice name
			name := args[0]

			// Get namespace option
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Execute command
			microservice := deploy.New()
			err = microservice.Execute(namespace, name)
			util.Check(err)
		},
	}

	return cmd
}
