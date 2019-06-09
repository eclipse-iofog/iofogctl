package cmd

import (
	"github.com/eclipse-iofog/iofogctl/internal/describe"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe resource NAME",
		Short: "Get detailed information of existing resources",
		Long:  `Get detailed information of existing resources`,
		Example: `iofogctl describe controller NAME
iofogctl describe agent NAME
iofogctl describe microservice NAME`,
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
