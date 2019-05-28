package deployagent

import (
	"github.com/eclipse-iofog/cli/pkg/util"
	"github.com/spf13/cobra"
)

type options struct {
	user    *string
	host    *string
	keyFile *string
	local   *bool
}

// NewCommand export
func NewCommand() *cobra.Command {
	// Instantiate options
	opt := &options{}

	// Instantiate command
	cmd := &cobra.Command{
		Use:   "agent name",
		Short: "Deploy an Agent",
		Long:  `Deploy an Agent`,
		Example: `iofog deploy agent my_agent_name --local
iofog deploy agent my_agent_name --user root --host 32.23.134.3 --key_file ~/.ssh/id_ecdsa`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get agent name and namespace
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Get executor to run procedure
			exe, err := getExecutor(opt)
			util.Check(err)

			// Execute
			err = exe.execute(namespace, name)
			util.Check(err)
		},
	}

	// Set up options
	opt.user = cmd.Flags().StringP("user", "u", "", "Username of host the Controller is being deployed on")
	opt.host = cmd.Flags().StringP("host", "o", "", "IP or hostname of host the Controller is being deployed on")
	opt.keyFile = cmd.Flags().StringP("key-file", "k", "", "Filename of SSH private key used to access host. Corresponding *.pub must be in same dir")
	opt.local = cmd.Flags().BoolP("local", "l", false, "Configure for local deployment. Cannot be used with other flags")
	cmd.Flags().Lookup("local").NoOptDefVal = "true"

	return cmd
}
