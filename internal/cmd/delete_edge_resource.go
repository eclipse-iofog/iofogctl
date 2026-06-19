package cmd

import (
	"fmt"

	delete "github.com/eclipse-iofog/iofogctl/internal/delete/edgeresource"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteEdgeResourceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edge-resource NAME VERSION",
		Short: "Delete an Edge Resource",
		Long: `Delete an Edge Resource.

Only the specified version will be deleted.
Agents that this Edge Resource are attached to will be notified of the deletion.`,
		Example: `iofogctl delete edge-resource NAME VERSION`,
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace of edge resource
			name := args[0]
			version := args[1]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Run the command
			exe := delete.NewExecutor(namespace, name, version)
			err = exe.Execute()
			util.Check(err)

			msg := fmt.Sprintf("Successfully deleted %s/%s", name, version)
			util.PrintSuccess(msg)
		},
	}

	return cmd
}
