package deploycontroller

import (
	"github.com/eclipse-iofog/cli/pkg/util"
	"github.com/spf13/cobra"
)

// NewCommand export
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "controller name",
		Short: "Deploy a Controller",
		Long: `Deploy a Controller`,
		Example: `iofog deploy controller my_controller_name --local
iofog deploy controller my_controller_name --user root --host 32.23.134.3 --key_file ~/.ssh/id_ecdsa
iofog deploy controller my_controller_name --kube-config ~/.kube/conf`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)
			println(namespace)
			
		},
	}

	cmd.Flags().StringP("user", "u", "", "Username of host the Controller is being deployed on")
	cmd.Flags().StringP("host", "o", "", "IP or hostname of host the Controller is being deployed on")
	cmd.Flags().StringP("key_file", "k", "", "Filename of SSH private key used to access host. Corresponding *.pub must be in same dir")
	
	cmd.Flags().StringP("kube-config", "q", "", "Filename of Kubernetes cluster config file. Cannot be used with other flags")

	cmd.Flags().BoolP("local", "l", false, "Configure for local deployment. Cannot be used with other flags")
	cmd.Flags().Lookup("local").NoOptDefVal = "true"

	return cmd
}