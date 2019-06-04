package cmd

import (
	deploy "github.com/eclipse-iofog/cli/internal/deploy/controller"
	"github.com/eclipse-iofog/cli/pkg/util"
	"github.com/spf13/cobra"
)

func newDeployControllerCommand() *cobra.Command {
	// Instantiate options
	opt := &deploy.Options{}

	// Instantiate command
	cmd := &cobra.Command{
		Use:   "controller name",
		Short: "Deploy a Controller",
		Long:  `Deploy a Controller`,
		Example: `iofogctl deploy controller my_controller_name --local --pass hgh3uj87hyy
iofogctl deploy controller my_controller_name --user root --host 32.23.134.3 --key_file ~/.ssh/id_ecdsa --pass hgh3uj87hyy
iofogctl deploy controller my_controller_name --kube-config ~/.kube/conf --pass hgh3uj87hyy`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			// Get name and namespace of controller
			opt.Name = args[0]
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			// Get executor for command
			ctrl, err := deploy.NewExecutor(opt)
			util.Check(err)

			// Execute the command
			err = ctrl.Execute()
			util.Check(err)
		},
	}

	// Set up options
	cmd.Flags().StringVarP(&opt.User, "user", "u", "", "Username of host the Controller is being deployed on")
	cmd.Flags().StringVarP(&opt.Host, "host", "o", "", "IP or hostname of host the Controller is being deployed on")
	cmd.Flags().StringVarP(&opt.KeyFile, "key-file", "k", "", "Filename of SSH private key used to access host. Corresponding *.pub must be in same dir")
	cmd.Flags().StringVarP(&opt.KubeConfig, "kube-config", "q", "", "Filename of Kubernetes cluster config file")
	cmd.Flags().BoolVarP(&opt.Local, "local", "l", false, "Configure for local deployment")
	cmd.Flags().Lookup("local").NoOptDefVal = "true"
	cmd.Flags().StringVarP(&opt.Password, "pass", "p", "", "Password for ioFog user registered in Controller")

	return cmd
}
