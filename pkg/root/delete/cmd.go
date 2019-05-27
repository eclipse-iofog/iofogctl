package delete

import (
	"github.com/spf13/cobra"
	"github.com/eclipse-iofog/cli/pkg/root/delete/controller"
	"github.com/eclipse-iofog/cli/pkg/root/delete/agent"
	"github.com/eclipse-iofog/cli/pkg/root/delete/microservice"
)

// NewCommand export
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete existing ioFog resources",
		Long: `Delete existing ioFog resources`,
	}

	cmd.AddCommand(deletecontroller.NewCommand())
	cmd.AddCommand(deleteagent.NewCommand())
	cmd.AddCommand(deletemicroservice.NewCommand())
	return cmd
}