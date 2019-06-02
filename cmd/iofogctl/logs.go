package main

import (
	"github.com/eclipse-iofog/cli/internal/logs"
	"github.com/eclipse-iofog/cli/pkg/util"
	"github.com/spf13/cobra"
)

func newLogsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs resource name",
		Short: "Get log contents of deployed resource",
		Long:  `Get log contents of deployed resource`,
		Example: `iofogctl logs controller my_controller_name
iofogctl logs agent my_agent_name
iofogctl logs microservice my_microservice_name`,
		Args: cobra.ExactValidArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Get Resource type and name
			resource := args[0]
			name := args[1]

			// Validate first argument
			if _, exists := resources[resource]; !exists {
				util.Check(util.NewNotFoundError(resource))
			}

			// Get namespace option
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Instantiate logs executor
			exe, err := logs.NewExecutor(resource, namespace, name)
			util.Check(err)

			// Run the logs command
			err = exe.Execute()
			util.Check(err)
		},
	}

	return cmd
}

// Values accepted in resource type argument
var resources = map[string]bool{
	"controller":   true,
	"agent":        true,
	"microservice": true,
}
