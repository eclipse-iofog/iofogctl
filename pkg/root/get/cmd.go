package get

import (
	"github.com/eclipse-iofog/cli/pkg/util"
	"github.com/spf13/cobra"
	"errors"
)

// NewCommand export
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get information of existing resources",
		Long: `Get information of existing resources`,
		Run: func(cmd *cobra.Command, args []string) {
			// Parse flags and arguments
			err := validate(cmd, args)
			util.Check(err)

			// Perform get for specified resource
			get := new()
			err = get.execute(args[0])
			util.Check(err)
		},
	}

	return cmd
}

var resources = map[string]bool {
	"controllers": true,
	"agents": true,
	"microservices": true,
}

func validate(cmd *cobra.Command, args []string) error {
	// Validation error output header
	header := `todo
`

	// Did the user provide a resource name?
	if len(args) < 1 {
		msg := "You must specify a resource name"
		return errors.New(header + msg)
	}

	// Did the user request an appropriate resource?
	if _, exists := resources[args[0]]; !exists {
		msg := "Unknown resource '" + args[0] + "'"
		return errors.New(header + msg)
	}
	return nil
}