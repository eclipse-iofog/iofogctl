package cmd

import (
	attach "github.com/eclipse-iofog/iofogctl/internal/attach/agent"
	detach "github.com/eclipse-iofog/iofogctl/internal/detach/agent"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newMoveAgentCommand() *cobra.Command {
	force := false
	cmd := &cobra.Command{
		Use:     "agent NAME DEST_NAMESPACE",
		Short:   "Move an Agent to another Namespace",
		Long:    `Move an Agent to another Namespace`,
		Example: `iofogctl move agent NAME DEST_NAMESPACE`,
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Get args
			name := args[0]
			destNamespace := args[1]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Detach
			exe := detach.NewExecutor(namespace, name, force)
			err = exe.Execute()
			util.Check(err)
			// Invalidate cache between Executor invocations
			if namespace == destNamespace {
				clientutil.InvalidateCache()
			}
			// Attach
			exe = attach.NewExecutor(&attach.Options{Name: name, Namespace: destNamespace})
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess(getMoveSuccessMessage("Agent", name, "Namespace", destNamespace))
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Move Agent even if it is running Microservices")

	return cmd
}
