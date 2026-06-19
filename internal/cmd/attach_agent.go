package cmd

import (
	attach "github.com/eclipse-iofog/iofogctl/internal/attach/agent"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newAttachAgentCommand() *cobra.Command {
	opt := attach.Options{}
	cmd := &cobra.Command{
		Use:   "agent NAME",
		Short: "Attach an Agent to an existing Namespace",
		Long: `Attach a detached Agent to an existing Namespace.

The Agent will be provisioned with the Controller within the Namespace.`,
		Example: `iofogctl attach agent NAME`,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace of agent
			opt.Name = args[0]
			var err error
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			// Run the command
			exe := attach.NewExecutor(&opt)
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully attached Agent " + opt.Name + " to namespace " + opt.Namespace)
		},
	}

	return cmd
}
