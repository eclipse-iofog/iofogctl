package deploycontroller

import (
	"github.com/eclipse-iofog/cli/pkg/util"
	"github.com/spf13/cobra"
)

type options struct {
	user       *string
	host       *string
	keyFile    *string
	kubeConfig *string
	local      *bool
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
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			ctrl, err := getExecutor(opt)
			util.Check(err)

			err = ctrl.execute(namespace, name)
			util.Check(err)
		},
	}

	// Set up options
	opt.user = cmd.Flags().StringP("user", "u", "", "Username of host the Controller is being deployed on")
	opt.host = cmd.Flags().StringP("host", "o", "", "IP or hostname of host the Controller is being deployed on")
	opt.keyFile = cmd.Flags().StringP("key_file", "k", "", "Filename of SSH private key used to access host. Corresponding *.pub must be in same dir")
	opt.kubeConfig = cmd.Flags().StringP("kube-config", "q", "", "Filename of Kubernetes cluster config file. Cannot be used with other flags")
	opt.local = cmd.Flags().BoolP("local", "l", false, "Configure for local deployment. Cannot be used with other flags")
	cmd.Flags().Lookup("local").NoOptDefVal = "true"

	return cmd
}
