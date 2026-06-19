package cmd

import (
	"github.com/spf13/cobra"
)

func newRenameCommand() *cobra.Command {
	// Instantiate command
	cmd := &cobra.Command{
		Use:   "rename",
		Short: "Rename the iofog resources that are currently deployed",
		Long:  `Rename the iofog resources that are currently deployed`,
	}

	// Add subcommands
	cmd.AddCommand(
		newRenameNamespaceCommand(),
		newRenameControllerCommand(),
		newRenameAgentCommand(),
		newRenameApplicationCommand(),
		newRenameMicroserviceCommand(),
		newRenameEdgeResourceCommand(),
	)

	return cmd
}
