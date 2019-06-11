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
		Use:   "connect CONTROLLERNAME",
		Short: "Connect to existing ioFog Controller and Agents",
		Long:  `Connect to existing ioFog Controller and Agents`,
		Example: `iofogctl connect CONTROLLERNAME --host 123.321.123.22:51121 --email EMAIL --pass PASSWORD
iofogctl connect CONTROLLERNAME --kube-config ~/.kube/conf --email EMAIL --pass PASSWORD`,
		Args: cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get resource name
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
	cmd.Flags().StringVarP(&opt.Endpoint, "endpoint", "o", "", "Host and port (host:port) of the Controller you are connecting to")
	cmd.Flags().StringVarP(&opt.KubeFile, "kube-config", "q", "", "Filename of Kubernetes cluster config file")
	cmd.Flags().StringVarP(&opt.Email, "email", "e", "", "Email address of user registered against Controller")
	cmd.Flags().StringVarP(&opt.Password, "pass", "p", "", "Password of user registered against Controller")

	return cmd
}
