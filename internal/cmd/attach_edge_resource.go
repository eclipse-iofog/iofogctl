package cmd

import (
	"fmt"

	attach "github.com/eclipse-iofog/iofogctl/internal/attach/edgeresource"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newAttachEdgeResourceCommand() *cobra.Command {
	opt := attach.Options{}
	cmd := &cobra.Command{
		Use:     "edge-resource NAME VERSION AGENT_NAME",
		Short:   "Attach an Edge Resource to an existing Agent",
		Long:    `Attach an Edge Resource to an existing Agent.`,
		Example: `iofogctl attach edge-resource NAME VERSION AGENT_NAME`,
		Args:    cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace of agent
			opt.Name = args[0]
			opt.Version = args[1]
			opt.Agent = args[2]
			var err error
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			// Run the command
			exe := attach.NewExecutor(opt)
			err = exe.Execute()
			util.Check(err)

			msg := fmt.Sprintf("Successfully attached EdgeResource %s/%s to Agent %s", opt.Name, opt.Version, opt.Agent)
			util.PrintSuccess(msg)
		},
	}

	return cmd
}
