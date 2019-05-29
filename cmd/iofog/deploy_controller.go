package main

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
		Example: `iofog deploy controller my_controller_name --local
iofog deploy controller my_controller_name --user root --host 32.23.134.3 --key_file ~/.ssh/id_ecdsa
iofog deploy controller my_controller_name --kube-config ~/.kube/conf`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace of controller
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Get executor for the procedure
			ctrl, err := deploy.NewExecutor(opt)
			util.Check(err)

			// Execute procedure
			err = ctrl.Execute(namespace, name)
			util.Check(err)
		},
	}

	// Set up options
	cmd.Flags().StringVarP(&opt.User, "user", "u", "", "Username of host the Controller is being deployed on")
	cmd.Flags().StringVarP(&opt.Host, "host", "o", "", "IP or hostname of host the Controller is being deployed on")
	cmd.Flags().StringVarP(&opt.KeyFile, "key-file", "k", "", "Filename of SSH private key used to access host. Corresponding *.pub must be in same dir")
	cmd.Flags().StringVarP(&opt.KubeConfig, "kube-config", "q", "", "Filename of Kubernetes cluster config file. Cannot be used with other flags")
	cmd.Flags().BoolVarP(&opt.Local, "local", "l", false, "Configure for local deployment. Cannot be used with other flags")
	cmd.Flags().Lookup("local").NoOptDefVal = "true"

	return cmd
}
