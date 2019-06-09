package cmd

import (
	"github.com/eclipse-iofog/iofogctl/internal/disconnect"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDisconnectCommand() *cobra.Command {
	//Instantiate options
	opt := &disconnect.Options{}

	// Instantiate command
	cmd := &cobra.Command{
		Use:   "disconnect CONTROLLERNAME",
		Short: "Disconnect from existing ioFog Controller and Agents",
		Long:  `Disconnect from existing ioFog Controller and Agents`,
		Example: `iofogctl disconnect CONTROLLERNAME
iofogctl disconnect CONTROLLERNAME`,
		Args: cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get Controller name
			opt.Name = args[0]

			// Get namespace option
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
