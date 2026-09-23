package cmd

import (
	"fmt"
	"strings"

	attach "github.com/eclipse-iofog/iofogctl/internal/attach/knowledge"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newAttachKnowledgeCommand() *cobra.Command {
	opt := attach.Options{}
	cmd := &cobra.Command{
		Use:     "knowledge NAME AGENT_NAME1 AGENT_NAME2",
		Short:   "Attach Knowledge to existing Agents",
		Long:    `Attach Knowledge to existing Agents.`,
		Example: ex(`%[1]s attach knowledge NAME AGENT_NAME1 AGENT_NAME2`),
		Args:    cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			opt.Name = args[0]
			opt.Agents = args[1:]
			var err error
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			exe := attach.NewExecutor(opt)
			err = exe.Execute()
			util.Check(err)

			msg := fmt.Sprintf("Successfully attached Knowledge %s to Agents %s", opt.Name, strings.Join(opt.Agents, ", "))
			util.PrintSuccess(msg)
		},
	}

	return cmd
}
