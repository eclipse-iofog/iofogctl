package cmd

import (
	move "github.com/eclipse-iofog/iofogctl/internal/move/microservice"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newMoveMicroserviceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "microservice NAME AGENT_NAME",
		Short:   "Move a Microservice to another Agent in the same Namespace",
		Long:    `Move a Microservice to another Agent in the same Namespace`,
		Example: `iofogctl move microservice NAME AGENT_NAME`,
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace
			name := args[0]
			agent := args[1]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Get an executor for the command
			err = move.Execute(namespace, name, agent)
			util.Check(err)

			util.PrintSuccess(getMoveSuccessMessage("Microservice", name, "Agent", agent))
		},
	}

	return cmd
}
