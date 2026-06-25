package cmd

import (
	delete "github.com/eclipse-iofog/iofogctl/internal/delete/agent"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteAgentCommand() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "agent NAME",
		Short: "Delete an Agent",
		Long: ex(`Delete an Agent.

The Agent will be unprovisioned from the Controller within the namespace.

The Agent stack will be uninstalled from the host.

If you wish to not remove the Agent stack from the host, please use %[1]s detach agent`),
		Example: ex(`%[1]s delete agent NAME`),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace of agent
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)
			useDetached, err := cmd.Flags().GetBool("detached")
			util.Check(err)

			// Run the command
			exe, err := delete.NewExecutor(namespace, name, useDetached, force)
			util.Check(err)
			err = exe.Execute()
			util.Check(err)

			printName := name
			if !useDetached {
				printName = namespace + "/" + name
			}
			util.PrintSuccess("Successfully deleted " + printName)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Remove even if there are still Microservices running on the Agent")
	cmd.Flags().Bool("detached", false, pkg.flagDescDetached)

	return cmd
}
