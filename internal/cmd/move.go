package cmd

import (
	"github.com/spf13/cobra"
)

func newMoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move",
		Short: "Move an existing resources inside the current Namespace",
		Long:  `Move an existing resources inside the current Namespace`,
	}

	cmd.AddCommand(
		newMoveMicroserviceCommand(),
		newMoveAgentCommand(),
	)

	return cmd
}
