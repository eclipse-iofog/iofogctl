package cmd

import (
	"github.com/eclipse-iofog/iofogctl/internal/connect"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newConnectCommand() *cobra.Command {
	//Instantiate options
	opt := &connect.Options{}

	// Instantiate command
	cmd := &cobra.Command{
		Use:   "connect controller_name",
		Short: "Connect to existing ioFog Controller and Agents",
		Long:  `Connect to existing ioFog Controller and Agents`,
		Example: `iofogctl connect my_controller_name --host 123.321.123.22
iofogctl connect my_controller_name --kube-config ~/.kube/conf`,
		Args: cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get Controller name
			opt.Name = args[0]

			// Get namespace option
			var err error
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			// Get executor for get command
			exe, err := connect.NewExecutor(opt)
			util.Check(err)

			// Execute the get command
			err = exe.Execute()
			util.Check(err)
		},
	}
	cmd.Flags().StringVarP(&opt.Host, "host", "o", "", "IP or hostname of host the Controller is being deployed on")
	cmd.Flags().StringVarP(&opt.KubeFile, "kube-config", "q", "", "Filename of Kubernetes cluster config file")

	return cmd
}
