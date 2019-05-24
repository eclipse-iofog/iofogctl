package logs

import (
	"github.com/spf13/cobra"
)

// NewCommand export
func NewCommand() *cobra.Command {
	impl := new()

	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Get logs of running Controller, Agent, or Microservice",
		Long: "Get logs of running Controller, Agent, or Microservice",
		Run: func(cmd *cobra.Command, args []string) {
			impl.validate(args)
			impl.execute()
		},
	}
	return cmd
}