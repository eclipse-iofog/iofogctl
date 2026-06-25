package cmd

import (
	"fmt"

	attachagent "github.com/eclipse-iofog/iofogctl/internal/attach/exec/agent"
	attach "github.com/eclipse-iofog/iofogctl/internal/attach/exec/microservice"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func NewAttachExecMicroserviceCommand() *cobra.Command {
	opt := attach.Options{}
	cmd := &cobra.Command{
		Use:     "microservice NAME",
		Short:   "Attach an Exec Session to a Microservice",
		Long:    `Attach an Exec Session to an existing Microservice.`,
		Example: ex(`%[1]s attach exec microservice AppName/MicroserviceName`),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			opt.Name = args[0]
			var err error
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			// Run the command
			exe := attach.NewExecutor(opt)
			err = exe.Execute()
			util.Check(err)

			msg := fmt.Sprintf("Successfully attached Exec Session to Microservice %s", opt.Name)
			util.PrintSuccess(msg)
		},
	}

	return cmd
}

func newAttachExecAgentCommand() *cobra.Command {
	opt := attachagent.Options{}
	cmd := &cobra.Command{
		Use:     "agent NAME [DEBUG_IMAGE]",
		Short:   "Attach an Exec Session to an Agent",
		Long:    `Attach an Exec Session to an existing Agent.`,
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

			// Run the command
			exe := attachagent.NewExecutor(opt)
			err = exe.Execute()
			util.Check(err)

			msg := fmt.Sprintf("Successfully attached Exec Session to Agent %s", opt.Name)
			util.PrintSuccess(msg)
		},
	}

	return cmd
}

func newAttachExecCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "exec",
		Short:   "Attach an Exec Session to a resource",
		Long:    `Attach an Exec Session to a Microservice or Agent.`,
		Example: ex(`%[1]s attach exec microservice AppName/MicroserviceName`),
	}

	// Add subcommands
	cmd.AddCommand(
		NewAttachExecMicroserviceCommand(),
		newAttachExecAgentCommand(),
	)

	return cmd
}
