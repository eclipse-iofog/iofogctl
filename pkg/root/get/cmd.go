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
		Long: `Get information of existing resources`,
		Example: `iofog get controllers
iofog get agents
iofog get microservices`,
		ValidArgs: []string{"controllers", "agents", "microservices"},
		Args: cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Perform get for specified resource
			get := new()
			err := get.execute(args[0])
			util.Check(err)
		},
	}

	return cmd
}