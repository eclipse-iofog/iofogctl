package main

import (
	"github.com/spf13/cobra"
)

func newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an ioFog resource",
		Long:  `Create an ioFog resource`,
	}

	// Add subcommands
	cmd.AddCommand(
		newCreateNamespaceCommand(),
	)
	return cmd
}
