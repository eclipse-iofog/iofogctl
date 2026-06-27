package cmd

import (
	"github.com/spf13/cobra"
)

func newStopCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stops a resource",
		Long:  "Stops a resource",
	}

	// Add subcommands
	cmd.AddCommand(
		newStopApplicationCommand(),
		newStopMicroserviceCommand(),
	)
	return cmd
}
