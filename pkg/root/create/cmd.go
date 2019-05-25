package create

import (
	"github.com/spf13/cobra"
	"github.com/eclipse-iofog/cli/pkg/root/create/controller"
)

// NewCommand export
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get information of existing resources",
		Long: `Get information of existing resources`,
		Run: func(cmd *cobra.Command, args []string) {
			println("HELLO")
			for _, val := range args {
				println(val)
			}
		},
	}

	cmd.AddCommand(controller.NewCommand())
	return cmd
}