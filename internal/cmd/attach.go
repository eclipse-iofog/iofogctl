package cmd

import (
	"github.com/spf13/cobra"
)

func newAttachCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "attach",
		Example: `attach`,
		Short:   "Attach one ioFog resource to another",
		Long:    `Attach one ioFog resource to another.`,
	}

	// Add subcommands
	cmd.AddCommand(
		newAttachAgentCommand(),
		newAttachEdgeResourceCommand(),
		newAttachVolumeMountCommand(),
		newAttachExecCommand(),
	)

	return cmd
}
