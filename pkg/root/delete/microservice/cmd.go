package deletemicroservice

import (
	"github.com/eclipse-iofog/cli/pkg/util"
	"github.com/spf13/cobra"
)

// NewCommand export
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "microservice name [agent]",
		Short: "Delete a Microservice",
		Long:  `Delete a Microservice`,
		Example: `iofog delete microservice my_microservice_name
iofog delete microservice my_microservice_name`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			microservice := new()
			err = microservice.execute(namespace, name)
			util.Check(err)
		},
	}

	return cmd
}
