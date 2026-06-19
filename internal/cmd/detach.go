package cmd

import (
	"github.com/spf13/cobra"
)

func newDetachCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "detach",
		Example: `detach`,
		Short:   "Detach one ioFog resource from another",
		Long:    `Detach one ioFog resource from another.`,
	}

	// Add subcommands
	cmd.AddCommand(
		newDetachAgentCommand(),
		newDetachEdgeResourceCommand(),
		newDetachVolumeMountCommand(),
		newDetachExecCommand(),
	)

	return cmd
}
