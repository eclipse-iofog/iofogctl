package cmd

import (
	"fmt"
	"strings"

	detach "github.com/eclipse-iofog/iofogctl/internal/detach/volumemount"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDetachVolumeMountCommand() *cobra.Command {
	opt := detach.Options{}
	cmd := &cobra.Command{
		Use:     "volume-mount NAME AGENT_NAME1 AGENT_NAME2",
		Short:   "Detach a Volume Mount from existing Agents",
		Long:    `Detach a Volume Mount from existing Agents.`,
		Example: `iofogctl detach volume-mount NAME AGENT_NAME1 AGENT_NAME2`,
		Args:    cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace of agent
			opt.Name = args[0]
			opt.Agents = args[1:]
			var err error
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			// Run the command
			exe := detach.NewExecutor(opt)
			err = exe.Execute()
			util.Check(err)

			msg := fmt.Sprintf("Successfully detached Volume Mount %s from Agents %s", opt.Name, strings.Join(opt.Agents, ", "))
			util.PrintSuccess(msg)
		},
	}

	return cmd
}
