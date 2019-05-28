package describe

import (
	"github.com/eclipse-iofog/cli/pkg/util"
	"github.com/spf13/cobra"
)

var resources = map[string]bool{
	"namespace":    true,
	"controller":   true,
	"agent":        true,
	"microservice": true,
}

// NewCommand export
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe resource name",
		Short: "Get detailed information of existing resources",
		Long:  `Get detailed information of existing resources`,
		Example: `iofog describe controller my_controller_name
iofog describe agent my_agent_name
iofog describe microservice my_microservice_name`,
		Args: cobra.ExactValidArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			resource := args[0]
			name := args[1]

			// Validate first argument
			if _, exists := resources[resource]; !exists {
				util.Check(util.NewNotFoundError(resource))
			}

			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			exe, err := getExecutor(resource)
			util.Check(err)

			err = exe.execute(namespace, name)
			util.Check(err)
		},
	}

	return cmd
}
