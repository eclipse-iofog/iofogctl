package cmd

import (
	"github.com/spf13/cobra"
)

func newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a resource",
		Long:  `Create a component of an Edge Compute Network.`,
	}

	// Add subcommands
	cmd.AddCommand(
		newCreateNamespaceCommand(),
	)
	return cmd
}
