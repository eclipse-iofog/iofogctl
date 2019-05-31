package main

import (
	"github.com/spf13/cobra"
)

func newDeployCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy ioFog stack on existing infrastructure",
		Long:  `Deploy ioFog stack on existing infrastructure`,
	}

	// Add subcommands
	cmd.AddCommand(
		newDeployControllerCommand(),
		newDeployAgentCommand(),
		newDeployMicroserviceCommand(),
	)
	return cmd
}
