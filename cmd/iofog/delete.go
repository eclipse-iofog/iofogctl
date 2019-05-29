package main

import (
	"github.com/spf13/cobra"
)

func newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete existing ioFog resources",
		Long:  `Delete existing ioFog resources`,
	}

	cmd.AddCommand(
		newDeleteControllerCommand(),
		newDeleteAgentCommand(),
		newDeleteMicroserviceCommand(),
	)
	return cmd
}
