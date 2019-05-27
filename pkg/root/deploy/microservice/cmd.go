package deploymicroservice

import (
	"github.com/eclipse-iofog/cli/pkg/util"
	"github.com/spf13/cobra"
)

// NewCommand export
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "microservice name [agent]",
		Short: "Deploy a Microservice",
		Long: `Deploy a Microservice`,
		Example: `iofog deploy microservice my_microservice_name
iofog deploy microservice my_microservice_name my_agent_name`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)
			println(namespace)
			
		},
	}

	return cmd
}