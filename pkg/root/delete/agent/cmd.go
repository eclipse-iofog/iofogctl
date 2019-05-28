package deleteagent

import (
	"github.com/eclipse-iofog/cli/pkg/util"
	"github.com/spf13/cobra"
)

// NewCommand export
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "agent name",
		Short:   "Delete an Agent",
		Long:    `Delete an Agent`,
		Example: `iofog delete agent my_agent_name`,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace of agent
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Get an executor for the procedure
			exe, err := getExecutor(namespace, name)
			util.Check(err)

			// Run the procedure
			err = exe.execute()
			util.Check(err)
		},
	}

	return cmd
}
