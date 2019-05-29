package main

import (
	deploy "github.com/eclipse-iofog/cli/internal/deploy/agent"
	"github.com/eclipse-iofog/cli/pkg/util"
	"github.com/spf13/cobra"
)

func newDeployAgentCommand() *cobra.Command {
	// Instantiate options
	opt := &deploy.Options{}

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
			exe, err := deploy.NewExecutor(opt)
			util.Check(err)

			// Execute
			err = exe.Execute(namespace, name)
			util.Check(err)
		},
	}

	// Set up options
	cmd.Flags().StringVarP(&opt.User, "user", "u", "", "Username of host the Controller is being deployed on")
	cmd.Flags().StringVarP(&opt.Host, "host", "o", "", "IP or hostname of host the Controller is being deployed on")
	cmd.Flags().StringVarP(&opt.KeyFile, "key-file", "k", "", "Filename of SSH private key used to access host. Corresponding *.pub must be in same dir")
	cmd.Flags().BoolVarP(&opt.Local, "local", "l", false, "Configure for local deployment. Cannot be used with other flags")
	cmd.Flags().Lookup("local").NoOptDefVal = "true"

	return cmd
}
