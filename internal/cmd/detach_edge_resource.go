package cmd

import (
	"fmt"

	detach "github.com/eclipse-iofog/iofogctl/internal/detach/edgeresource"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDetachEdgeResourceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "edge-resource NAME VERSION AGENT_NAME",
		Short:   "Detaches an Edge Resource from an Agent",
		Long:    `Detaches an Edge Resource from an Agent.`,
		Example: `iofogctl detach edge-resource NAME VERSION AGENT_NAME`,
		Args:    cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace of edge resource
			name := args[0]
			version := args[1]
			agent := args[2]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Run the command
			exe := detach.NewExecutor(namespace, name, version, agent)
			err = exe.Execute()
			util.Check(err)

			msg := fmt.Sprintf("Successfully detached %s/%s", name, version)
			util.PrintSuccess(msg)
		},
	}

	return cmd
}
