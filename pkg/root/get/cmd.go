package get

import (
	"github.com/eclipse-iofog/cli/pkg/util"
	"github.com/spf13/cobra"
)

// NewCommand export
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get resource",
		Short: "Get information of existing resources",
		Long:  `Get information of existing resources`,
		Example: `iofog get controllers
iofog get agents
iofog get microservices`,
		ValidArgs: []string{"namespaces", "controllers", "agents", "microservices"},
		Args:      cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Perform get for specified resource
			resource := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			exe, err := getExecutor(resource)
			util.Check(err)

			err = exe.execute(namespace)
			util.Check(err)
		},
	}

	return cmd
}
