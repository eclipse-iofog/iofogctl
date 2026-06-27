package cmd

import (
	"fmt"

	attachagent "github.com/eclipse-iofog/iofogctl/internal/attach/exec/agent"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newAttachExecAgentCommand() *cobra.Command {
	opt := attachagent.Options{}
	cmd := &cobra.Command{
		Use:     "agent NAME [DEBUG_IMAGE]",
		Short:   "Provision a fog debug exec microservice on an Agent",
		Long:    `Provision a debug microservice on an Agent for interactive exec via POST /iofog/{uuid}/exec.`,
		Example: ex(`%[1]s attach exec agent AgentName DebugImage`),
		Args:    cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			opt.Name = args[0]
			if len(args) > 1 {
				opt.Image = &args[1]
			}
			var err error
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			exe := attachagent.NewExecutor(opt)
			err = exe.Execute()
			util.Check(err)

			msg := fmt.Sprintf("Successfully provisioned debug exec for Agent %s", opt.Name)
			util.PrintSuccess(msg)
		},
	}

	return cmd
}

func newAttachExecCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "exec",
		Short:   "Provision fog debug exec on an Agent",
		Long:    `Provision fog debug exec resources. Use exec agent to open an interactive shell after provisioning.`,
		Example: ex(`%[1]s attach exec agent AgentName`),
	}

	cmd.AddCommand(newAttachExecAgentCommand())

	return cmd
}
