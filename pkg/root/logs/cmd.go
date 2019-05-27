package logs

import (
	"github.com/eclipse-iofog/cli/pkg/util"
	"github.com/spf13/cobra"
)

var resources = map[string]bool{
	"controller": true,
	"agent": true,
	"microservice": true,
}

// NewCommand export
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs resource name",
		Short: "Get log contents of deployed resource",
		Long: `Get log contents of deployed resource`,
		Example: `iofog logs controller my_controller_name
iofog logs agent my_agent_name
iofog logs microservice my_microservice_name`,
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