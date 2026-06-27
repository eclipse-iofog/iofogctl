package cmd

import (
	"github.com/spf13/cobra"
)

func newExecCommand() *cobra.Command {
	// Values accepted in resource type argument
	cmd := &cobra.Command{
		Use:   "exec",
		Short: "Connect to an Exec Session of a resource",
		Long:  `Connect to an Exec Session of a Microservice or Agent.`,
	}

	// Add subcommands
	cmd.AddCommand(
		newExecMicroserviceCommand(),
		newExecAgentCommand(),
	)

	return cmd
}
