package cmd

import (
	"github.com/eclipse-iofog/iofogctl/internal/deploy"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeployCommand() *cobra.Command {
	// Instantiate options
	opt := &deploy.Options{}

	// Instantiate command
	cmd := &cobra.Command{
		Use: "deploy",
		Example: `deploy -f platform.yaml
deploy [command]`,
		Short: "Deploy ioFog stack on existing infrastructure",
		Long:  `Deploy ioFog stack on existing infrastructure`,
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			// Get namespace
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			// Execute command
			err = deploy.Execute(opt)
			util.Check(err)
		},
	}

	// Add subcommands
	cmd.AddCommand(
		newDeployControllerCommand(),
		newDeployAgentCommand(),
		newDeployMicroserviceCommand(),
	)

	// Register flags
	cmd.Flags().StringVarP(&opt.Filename, "file", "f", "", "YAML file containing resource definitions for Controllers, Agents, and Microservice to deploy")

	return cmd
}
