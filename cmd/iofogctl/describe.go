package main

import (
	"github.com/eclipse-iofog/cli/internal/describe"
	"github.com/eclipse-iofog/cli/pkg/util"
	"github.com/spf13/cobra"
)

func newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe resource name",
		Short: "Get detailed information of existing resources",
		Long:  `Get detailed information of existing resources`,
		Example: `iofogctl describe controller my_controller_name
iofogctl describe agent my_agent_name
iofogctl describe microservice my_microservice_name`,
		Args: cobra.ExactValidArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Get resource type and name
			resource := args[0]
			name := args[1]

			// Get namespace option
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Validate first argument
			if _, exists := resources[resource]; !exists {
				util.Check(util.NewNotFoundError(resource))
			}

			// Get executor for describe command
			exe, err := describe.NewExecutor(resource, namespace, name)
			util.Check(err)

			// Execute the command
			err = exe.Execute()
			util.Check(err)
		},
	}

	return cmd
}
