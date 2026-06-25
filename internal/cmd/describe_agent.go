package cmd

import (
	"github.com/eclipse-iofog/iofogctl/internal/describe"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDescribeAgentCommand() *cobra.Command {
	opt := describe.Options{
		Resource: "agent",
	}

	cmd := &cobra.Command{
		Use:     "agent NAME",
		Short:   "Get detailed information about an Agent",
		Long:    `Get detailed information about a named Agent.`,
		Example: ex(`%[1]s describe agent NAME`),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get resource type and name
			var err error
			opt.Name = args[0]
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			// Get executor for describe command
			exe, err := describe.NewExecutor(&opt)
			util.Check(err)

			// Execute the command
			err = exe.Execute()
			util.Check(err)
		},
	}
	cmd.Flags().StringVarP(&opt.Filename, "output-file", "o", "", "YAML output file")
	cmd.Flags().BoolVarP(&opt.IsDetached, "detached", "", false, pkg.flagDescDetached)

	return cmd
}
