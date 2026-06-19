package cmd

import (
	rename "github.com/eclipse-iofog/iofogctl/internal/rename/edgeresource"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newRenameEdgeResourceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "edge-resource NAME NEW_NAME",
		Short:   "Rename an Edge Resource",
		Long:    `Rename an Edge Resource`,
		Example: `iofogctl rename edge-resource NAME NEW_NAME`,
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and new name of the edgeResource
			name := args[0]
			newName := args[1]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Get an executor for the command
			err = rename.Execute(namespace, name, newName)
			util.Check(err)

			util.PrintSuccess(getRenameSuccessMessage("Edge Resource", name, newName))
		},
	}

	return cmd
}
