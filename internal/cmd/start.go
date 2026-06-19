package cmd

import (
	"github.com/spf13/cobra"
)

func newStartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Starts a resource",
		Long:  "Starts a resource",
	}

	// Add subcommands
	cmd.AddCommand(
		newStartApplicationCommand(),
		newStartMicroserviceCommand(),
	)
	return cmd
}
