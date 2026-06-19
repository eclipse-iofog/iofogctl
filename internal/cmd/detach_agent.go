package cmd

import (
	detach "github.com/eclipse-iofog/iofogctl/internal/detach/agent"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDetachAgentCommand() *cobra.Command {
	force := false
	cmd := &cobra.Command{
		Use:   "agent NAME",
		Short: "Detaches an Agent",
		Long: `Detaches an Agent.

The Agent will be deprovisioned from the Controller within the namespace.
The Agent will be removed from Controller.

You cannot detach unprovisioned Agents.

The Agent stack will not be uninstalled from the host.`,
		Example: `iofogctl detach agent NAME`,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace of agent
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Run the command
			exe := detach.NewExecutor(namespace, name, force)
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully detached " + name)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Detach Agent even if it is running Microservices")

	return cmd
}
