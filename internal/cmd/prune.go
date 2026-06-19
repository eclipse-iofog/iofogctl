package cmd

import (
	"github.com/spf13/cobra"
)

func newDockerPruneCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prune",
		Short: "prune ioFog resources",
		Long:  `prune ioFog resources`,
	}

	// Add subcommands
	cmd.AddCommand(
		newPruneAgentCommand(),
	)

	return cmd
}
