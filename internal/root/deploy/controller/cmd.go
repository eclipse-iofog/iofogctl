package deploycontroller

import (
	"github.com/eclipse-iofog/cli/pkg/util"
	"github.com/spf13/cobra"
)

type options struct {
	user       string
	host       string
	keyFile    string
	kubeConfig string
	local      bool
}

// NewCommand export
func NewCommand() *cobra.Command {
	// Instantiate options
	opt := &options{}

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
			ctrl, err := getExecutor(opt)
			util.Check(err)

			// Execute procedure
			err = ctrl.execute(namespace, name)
			util.Check(err)
		},
	}

	// Set up options
	cmd.Flags().StringVarP(&opt.user, "user", "u", "", "Username of host the Controller is being deployed on")
	cmd.Flags().StringVarP(&opt.host, "host", "o", "", "IP or hostname of host the Controller is being deployed on")
	cmd.Flags().StringVarP(&opt.keyFile, "key-file", "k", "", "Filename of SSH private key used to access host. Corresponding *.pub must be in same dir")
	cmd.Flags().StringVarP(&opt.kubeConfig, "kube-config", "q", "", "Filename of Kubernetes cluster config file. Cannot be used with other flags")
	cmd.Flags().BoolVarP(&opt.local, "local", "l", false, "Configure for local deployment. Cannot be used with other flags")
	cmd.Flags().Lookup("local").NoOptDefVal = "true"

	return cmd
}
