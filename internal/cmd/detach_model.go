package cmd

import (
	"fmt"
	"strings"

	detach "github.com/eclipse-iofog/iofogctl/internal/detach/model"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDetachModelCommand() *cobra.Command {
	opt := detach.Options{}
	cmd := &cobra.Command{
		Use:     "model NAME AGENT_NAME1 AGENT_NAME2",
		Short:   "Detach a Model from existing Agents",
		Long:    `Detach a Model from existing Agents.`,
		Example: ex(`%[1]s detach model NAME AGENT_NAME1 AGENT_NAME2`),
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

			msg := fmt.Sprintf("Successfully detached Model %s from Agents %s", opt.Name, strings.Join(opt.Agents, ", "))
			util.PrintSuccess(msg)
		},
	}

	return cmd
}
