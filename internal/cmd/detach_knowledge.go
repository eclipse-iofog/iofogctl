package cmd

import (
	"fmt"
	"strings"

	detach "github.com/eclipse-iofog/iofogctl/internal/detach/knowledge"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDetachKnowledgeCommand() *cobra.Command {
	opt := detach.Options{}
	cmd := &cobra.Command{
		Use:     "knowledge NAME AGENT_NAME1 AGENT_NAME2",
		Short:   "Detach Knowledge from existing Agents",
		Long:    `Detach Knowledge from existing Agents.`,
		Example: ex(`%[1]s detach knowledge NAME AGENT_NAME1 AGENT_NAME2`),
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

			msg := fmt.Sprintf("Successfully detached Knowledge %s from Agents %s", opt.Name, strings.Join(opt.Agents, ", "))
			util.PrintSuccess(msg)
		},
	}

	return cmd
}
