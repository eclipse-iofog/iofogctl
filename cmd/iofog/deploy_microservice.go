package main

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
		Example: `iofog deploy microservice my_microservice_name
iofog deploy microservice my_microservice_name my_agent_name`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			microservice := deploy.New()
			err = microservice.Execute(namespace, name)
			util.Check(err)
		},
	}

	return cmd
}
