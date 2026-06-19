package cmd

import (
	"github.com/spf13/cobra"
)

func newRebuildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rebuild",
		Short: "Rebuilds a microservice or system-microservice",
		Long:  "Rebuilds a microservice or system-microservice",
	}

	// Add subcommands
	cmd.AddCommand(
		newRebuildMicroserviceCommand(),
		newRebuildSystemMicroserviceCommand(),
	)
	return cmd
}
