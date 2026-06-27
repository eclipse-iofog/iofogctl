package cmd

import (
	"github.com/eclipse-iofog/iofogctl/internal/exec"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newExecAgentCommand() *cobra.Command {
	opt := exec.Options{
		Resource: "agent",
	}

	cmd := &cobra.Command{
		Use:   "agent AgentName [DEBUG_IMAGE]",
		Short: "Open an interactive exec session on an Agent debug shell",
		Long:  `Open a WebSocket exec session to the Agent debug microservice. Provisions fog debug exec automatically when it is not already enabled.`,
		Example: ex(`%[1]s exec agent AgentName
%[1]s exec agent AgentName ghcr.io/org/debug:latest`),
		Args: cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			opt.Name = args[0]
			if len(args) > 1 {
				opt.DebugImage = &args[1]
			}
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			// Get executor for exec command
			exe, err := exec.NewExecutor(&opt)
			util.Check(err)

			// Execute the command
			err = exe.Execute()
			util.Check(err)
		},
	}

	return cmd
}
