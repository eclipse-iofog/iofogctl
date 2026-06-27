package cmd

import (
	"fmt"

	detachagent "github.com/eclipse-iofog/iofogctl/internal/detach/exec/agent"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDetachExecAgentCommand() *cobra.Command {
	opt := detachagent.Options{}
	cmd := &cobra.Command{
		Use:     "agent NAME",
		Short:   "Remove fog debug exec from an Agent",
		Long:    `Remove the debug microservice provisioned for Agent exec via DELETE /iofog/{uuid}/exec.`,
		Example: ex(`%[1]s detach exec agent AgentName`),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			opt.Name = args[0]
			var err error
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			exe := detachagent.NewExecutor(opt)
			err = exe.Execute()
			util.Check(err)

			msg := fmt.Sprintf("Successfully removed debug exec from Agent %s", opt.Name)
			util.PrintSuccess(msg)
		},
	}

	return cmd
}

func newDetachExecCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "exec",
		Short:   "Remove fog debug exec from an Agent",
		Long:    `Remove fog debug exec resources provisioned with attach exec agent.`,
		Example: ex(`%[1]s detach exec agent AgentName`),
	}

	cmd.AddCommand(newDetachExecAgentCommand())

	return cmd
}
