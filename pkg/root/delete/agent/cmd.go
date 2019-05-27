package deleteagent

import (
	"github.com/eclipse-iofog/cli/pkg/util"
	"github.com/spf13/cobra"
)

// NewCommand export
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent name",
		Short: "Delete an Agent",
		Long: `Delete an Agent`,
		Example: `iofog delete agent my_agent_name`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)
			println(namespace)
			
		},
	}

	return cmd
}