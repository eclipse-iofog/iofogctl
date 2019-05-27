package deployagent

import (
	"github.com/eclipse-iofog/cli/pkg/util"
	"github.com/spf13/cobra"
)

// NewCommand export
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent name",
		Short: "Deploy an Agent",
		Long: `Deploy an Agent`,
		Example: `iofog deploy agent my_agent_name --local
iofog deploy agent my_agent_name --user root --host 32.23.134.3 --key_file ~/.ssh/id_ecdsa`,
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
	
	cmd.Flags().BoolP("local", "l", false, "Configure for local deployment. Cannot be used with other flags")
	cmd.Flags().Lookup("local").NoOptDefVal = "true"

	return cmd
}