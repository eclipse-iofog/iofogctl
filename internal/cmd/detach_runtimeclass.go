package cmd

import (
	"fmt"
	"strings"

	detach "github.com/eclipse-iofog/iofogctl/internal/detach/runtimeclass"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDetachRuntimeClassCommand() *cobra.Command {
	opt := detach.Options{}
	cmd := &cobra.Command{
		Use:     "runtimeclass NAME AGENT_NAME1 AGENT_NAME2",
		Short:   "Detach a RuntimeClass from existing Agents",
		Long:    `Detach a RuntimeClass from existing Agents.`,
		Example: ex(`%[1]s detach runtimeclass NAME AGENT_NAME1 AGENT_NAME2`),
		Args:    cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			opt.Name = args[0]
			opt.Agents = args[1:]
			var err error
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			exe := detach.NewExecutor(opt)
			err = exe.Execute()
			util.Check(err)

			msg := fmt.Sprintf("Successfully detached RuntimeClass %s from Agents %s", opt.Name, strings.Join(opt.Agents, ", "))
			util.PrintSuccess(msg)
		},
	}

	return cmd
}
