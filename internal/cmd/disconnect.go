package cmd

import (
	"github.com/eclipse-iofog/iofogctl/internal/disconnect"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDisconnectCommand() *cobra.Command {
	// Instantiate options
	opt := &disconnect.Options{}

	// Instantiate command
	cmd := &cobra.Command{
		Use:   "disconnect",
		Short: "Disconnect from an ioFog cluster",
		Long: `Disconnect from an ioFog cluster.

This will remove all client-side information for this Namespace. The Namespace will itself be deleted.
Use the connect command to reconnect after a disconnect.
If you would like to uninstall the Control Plane and/or Agents, use the delete command instead.`,
		Example: `iofogctl disconnect -n NAMESPACE`,
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			// Execute the get command
			err = disconnect.Execute(opt)
			util.Check(err)
		},
	}

	return cmd
}
