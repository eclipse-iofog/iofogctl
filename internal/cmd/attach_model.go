package cmd

import (
	"fmt"
	"strings"

	attach "github.com/eclipse-iofog/iofogctl/internal/attach/model"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newAttachModelCommand() *cobra.Command {
	opt := attach.Options{}
	cmd := &cobra.Command{
		Use:     "model NAME AGENT_NAME1 AGENT_NAME2",
		Short:   "Attach a Model to existing Agents",
		Long:    `Attach a Model to existing Agents.`,
		Example: ex(`%[1]s attach model NAME AGENT_NAME1 AGENT_NAME2`),
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

			msg := fmt.Sprintf("Successfully attached Model %s to Agents %s", opt.Name, strings.Join(opt.Agents, ", "))
			util.PrintSuccess(msg)
		},
	}

	return cmd
}
